package skill

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// CastManager handles skill casting: validation, cooldowns, cast flow, and effect application.
// Uses callback injection to avoid import cycles with gameserver package.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: Creature.useMagic(), RequestMagicSkillUse.java
type CastManager struct {
	// cooldowns tracks skill cooldown expiry: key = "objectID_skillID"
	cooldowns sync.Map

	// sendPacketFunc sends a packet to a specific player (by objectID).
	sendPacketFunc func(objectID uint32, data []byte, size int)

	// broadcastFunc broadcasts to visible players around the source player.
	broadcastFunc func(source *model.Player, data []byte, size int)

	// getEffectManager returns the EffectManager for a character.
	getEffectManager func(objectID uint32) *EffectManager
}

// NewCastManager creates a new CastManager with injected callbacks.
func NewCastManager(
	sendPacketFunc func(objectID uint32, data []byte, size int),
	broadcastFunc func(source *model.Player, data []byte, size int),
	getEffectManager func(objectID uint32) *EffectManager,
) *CastManager {
	return &CastManager{
		sendPacketFunc:   sendPacketFunc,
		broadcastFunc:    broadcastFunc,
		getEffectManager: getEffectManager,
	}
}

// UseMagic handles a skill use request from a player.
// Validates skill, consumes MP, broadcasts cast animation, schedules effect application.
func (cm *CastManager) UseMagic(caster *model.Player, skillID int32, ctrl, shift bool) error {
	// 1. Check player has skill
	skillInfo := caster.GetSkill(skillID)
	if skillInfo == nil {
		return fmt.Errorf("skill %d not learned", skillID)
	}

	// 2. Get skill template
	tmpl := data.GetSkillTemplate(skillID, skillInfo.Level)
	if tmpl == nil {
		return fmt.Errorf("skill template not found: %d L%d", skillID, skillInfo.Level)
	}

	// 3. Passive skills cannot be cast
	if tmpl.IsPassive() {
		return fmt.Errorf("cannot cast passive skill %d", skillID)
	}

	// 4. Check cooldown
	cdKey := cooldownKey(caster.ObjectID(), skillID)
	if expiry, ok := cm.cooldowns.Load(cdKey); ok {
		if time.Now().Before(expiry.(time.Time)) {
			return fmt.Errorf("skill %d on cooldown", skillID)
		}
		cm.cooldowns.Delete(cdKey)
	}

	// 5. Check MP
	if tmpl.MpConsume > 0 && caster.CurrentMP() < tmpl.MpConsume {
		return fmt.Errorf("not enough MP: need %d, have %d", tmpl.MpConsume, caster.CurrentMP())
	}

	// 6. Check if dead
	if caster.IsDead() {
		return fmt.Errorf("cannot cast while dead")
	}

	// 7. Resolve target
	targetObjID := cm.resolveTarget(caster, tmpl)

	// 8. Consume MP
	if tmpl.MpConsume > 0 {
		caster.SetCurrentMP(caster.CurrentMP() - tmpl.MpConsume)
	}
	if tmpl.HpConsume > 0 {
		caster.SetCurrentHP(caster.CurrentHP() - tmpl.HpConsume)
	}

	// 9. Broadcast MagicSkillUse (cast animation)
	loc := caster.Location()
	msu := serverpackets.NewMagicSkillUse(
		int32(caster.ObjectID()),
		int32(targetObjID),
		skillID,
		skillInfo.Level,
		tmpl.HitTime,
		tmpl.ReuseDelay,
		loc.X, loc.Y, loc.Z,
	)
	msuData, err := msu.Write()
	if err != nil {
		return fmt.Errorf("writing MagicSkillUse: %w", err)
	}
	cm.broadcastFunc(caster, msuData, len(msuData))
	cm.sendPacketFunc(caster.ObjectID(), msuData, len(msuData))

	// 10. Schedule cast completion
	if tmpl.HitTime > 0 {
		// Delayed cast
		go cm.finishCast(caster, targetObjID, tmpl, skillInfo.Level)
	} else {
		// Instant cast
		cm.applyEffects(caster, targetObjID, tmpl, skillInfo.Level)
		cm.broadcastLaunched(caster, targetObjID, tmpl, skillInfo.Level)
	}

	// 11. Start cooldown
	if tmpl.ReuseDelay > 0 {
		cm.cooldowns.Store(cdKey, time.Now().Add(time.Duration(tmpl.ReuseDelay)*time.Millisecond))
	}

	slog.Debug("skill cast",
		"caster", caster.Name(),
		"skill", tmpl.Name,
		"skillID", skillID,
		"level", skillInfo.Level,
		"target", targetObjID,
		"hitTime", tmpl.HitTime)

	return nil
}

// IsOnCooldown checks if a skill is on cooldown for a player.
func (cm *CastManager) IsOnCooldown(objectID uint32, skillID int32) bool {
	cdKey := cooldownKey(objectID, skillID)
	if expiry, ok := cm.cooldowns.Load(cdKey); ok {
		return time.Now().Before(expiry.(time.Time))
	}
	return false
}

// finishCast waits for HitTime then applies effects.
func (cm *CastManager) finishCast(caster *model.Player, targetObjID uint32, tmpl *data.SkillTemplate, level int32) {
	time.Sleep(time.Duration(tmpl.HitTime) * time.Millisecond)

	// Check if caster is still alive
	if caster.IsDead() {
		return
	}

	cm.applyEffects(caster, targetObjID, tmpl, level)
	cm.broadcastLaunched(caster, targetObjID, tmpl, level)
}

// applyEffects applies all skill effects to the target.
func (cm *CastManager) applyEffects(caster *model.Player, targetObjID uint32, tmpl *data.SkillTemplate, level int32) {
	for _, effectDef := range tmpl.Effects {
		effect, err := CreateEffect(effectDef.Name, effectDef.Params)
		if err != nil {
			slog.Warn("failed to create effect",
				"effect", effectDef.Name,
				"skill", tmpl.Name,
				"error", err)
			continue
		}

		if effect.IsInstant() {
			effect.OnStart(caster.ObjectID(), targetObjID)
			continue
		}

		// Continuous effect — add to target's EffectManager
		ae := &ActiveEffect{
			CasterObjID:   caster.ObjectID(),
			TargetObjID:   targetObjID,
			SkillID:       tmpl.ID,
			SkillLevel:    level,
			Effect:        effect,
			RemainingMs:   tmpl.AbnormalTime * 1000, // convert seconds to ms
			AbnormalType:  tmpl.AbnormalType,
			AbnormalLevel: tmpl.AbnormalLevel,
		}

		if cm.getEffectManager != nil {
			if em := cm.getEffectManager(targetObjID); em != nil {
				// Determine if buff or debuff based on effect type
				if isBeneficial(effect) {
					em.AddBuff(ae)
				} else {
					em.AddDebuff(ae)
				}
			}
		}
	}
}

// broadcastLaunched sends MagicSkillLaunched to nearby players.
func (cm *CastManager) broadcastLaunched(caster *model.Player, targetObjID uint32, tmpl *data.SkillTemplate, level int32) {
	msl := serverpackets.NewMagicSkillLaunched(
		int32(caster.ObjectID()),
		tmpl.ID,
		level,
		[]int32{int32(targetObjID)},
	)
	mslData, err := msl.Write()
	if err != nil {
		slog.Error("failed to write MagicSkillLaunched",
			"skill", tmpl.Name,
			"error", err)
		return
	}
	cm.broadcastFunc(caster, mslData, len(mslData))
	cm.sendPacketFunc(caster.ObjectID(), mslData, len(mslData))
}

// resolveTarget determines the target objectID based on skill TargetType.
func (cm *CastManager) resolveTarget(caster *model.Player, tmpl *data.SkillTemplate) uint32 {
	switch tmpl.TargetType {
	case data.TargetSelf:
		return caster.ObjectID()
	case data.TargetOne:
		if target := caster.Target(); target != nil {
			return target.ObjectID()
		}
		return caster.ObjectID() // Fallback to self
	default:
		return caster.ObjectID()
	}
}

// isBeneficial returns true if the effect is beneficial (buff, heal).
func isBeneficial(e Effect) bool {
	switch e.Name() {
	case "Buff", "Heal", "MpHeal", "HealOverTime", "SpeedChange", "StatUp":
		return true
	default:
		return false
	}
}

// cooldownKey generates a unique key for cooldown tracking.
func cooldownKey(objectID uint32, skillID int32) string {
	return fmt.Sprintf("%d_%d", objectID, skillID)
}

// CastMgr is the global CastManager instance, initialized in main.go.
// Phase 5.9.4: Cast Flow & Packets.
var CastMgr *CastManager
