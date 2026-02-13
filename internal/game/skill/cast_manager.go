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

// activeCast tracks an in-progress skill cast for interrupt support.
// One activeCast per caster at a time (can't cast two skills simultaneously).
type activeCast struct {
	cancel  chan struct{} // closed to interrupt the cast
	skillID int32
}

// CastManager handles skill casting: validation, cooldowns, cast flow, and effect application.
// Uses callback injection to avoid import cycles with gameserver package.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: Creature.useMagic(), RequestMagicSkillUse.java
type CastManager struct {
	// cooldowns tracks skill cooldown expiry: key = "objectID_skillID"
	cooldowns sync.Map

	// activeCasts tracks in-progress casts: key = objectID, value = *activeCast
	activeCasts sync.Map

	// sendPacketFunc sends a packet to a specific player (by objectID).
	sendPacketFunc func(objectID uint32, data []byte, size int)

	// broadcastFunc broadcasts to visible players around the source player.
	broadcastFunc func(source *model.Player, data []byte, size int)

	// getEffectManager returns the EffectManager for a character.
	getEffectManager func(objectID uint32) *EffectManager

	// getWorldObject resolves an objectID to a WorldObject (for target revalidation).
	// Nil if not wired — target revalidation is skipped.
	getWorldObject func(objectID uint32) (*model.WorldObject, bool)
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

// SetWorldObjectResolver sets the callback for resolving objectID → WorldObject.
// Called after construction to avoid circular dependency at init time.
// Also sets the package-level resolver so effects can access game objects.
func (cm *CastManager) SetWorldObjectResolver(fn func(objectID uint32) (*model.WorldObject, bool)) {
	cm.getWorldObject = fn
	SetWorldResolver(fn)
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

	// 4. Check if already casting
	if caster.IsCasting() {
		return fmt.Errorf("already casting")
	}

	// 5. Check cooldown
	cdKey := cooldownKey(caster.ObjectID(), skillID)
	if expiry, ok := cm.cooldowns.Load(cdKey); ok {
		if time.Now().Before(expiry.(time.Time)) {
			return fmt.Errorf("skill %d on cooldown", skillID)
		}
		cm.cooldowns.Delete(cdKey)
	}

	// 6. Check MP
	if tmpl.MpConsume > 0 && caster.CurrentMP() < tmpl.MpConsume {
		return fmt.Errorf("not enough MP: need %d, have %d", tmpl.MpConsume, caster.CurrentMP())
	}

	// 7. Check if dead
	if caster.IsDead() {
		return fmt.Errorf("cannot cast while dead")
	}

	// 8. Resolve target
	targetObjID := cm.resolveTarget(caster, tmpl)

	// 9. Range check (skip for self-targeting skills)
	if tmpl.CastRange > 0 && targetObjID != caster.ObjectID() {
		if err := cm.checkRange(caster, targetObjID, tmpl.CastRange); err != nil {
			return err
		}
	}

	// 10. Consume MP/HP
	if tmpl.MpConsume > 0 {
		caster.SetCurrentMP(caster.CurrentMP() - tmpl.MpConsume)
	}
	if tmpl.HpConsume > 0 {
		caster.SetCurrentHP(caster.CurrentHP() - tmpl.HpConsume)
	}

	// 11. Broadcast MagicSkillUse (cast animation + cast bar)
	loc := caster.Location()
	targetLoc := loc
	if target := caster.Target(); target != nil && target.ObjectID() == targetObjID {
		targetLoc = target.Location()
	}
	msu := serverpackets.NewMagicSkillUse(
		int32(caster.ObjectID()),
		int32(targetObjID),
		skillID,
		skillInfo.Level,
		tmpl.HitTime,
		tmpl.ReuseDelay,
		loc.X, loc.Y, loc.Z,
		targetLoc.X, targetLoc.Y, targetLoc.Z,
		false, // critical
	)
	msuData, err := msu.Write()
	if err != nil {
		return fmt.Errorf("writing MagicSkillUse: %w", err)
	}
	cm.broadcastFunc(caster, msuData, len(msuData))
	cm.sendPacketFunc(caster.ObjectID(), msuData, len(msuData))

	// 12. Schedule cast completion
	if tmpl.HitTime > 0 {
		// Set casting flag and register active cast for interrupt support
		caster.SetCasting(true)
		ac := &activeCast{
			cancel:  make(chan struct{}),
			skillID: skillID,
		}
		cm.activeCasts.Store(caster.ObjectID(), ac)

		go cm.finishCast(caster, targetObjID, tmpl, skillInfo.Level, ac)
	} else {
		// Instant cast — apply immediately
		cm.applyEffects(caster, targetObjID, tmpl, skillInfo.Level)
		cm.broadcastLaunched(caster, targetObjID, tmpl, skillInfo.Level)
	}

	// 13. Start cooldown
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

// InterruptCast interrupts an active cast for the given objectID.
// Called when the character takes damage during a non-instant cast.
// Java reference: Creature.onHit() → abortCast().
func (cm *CastManager) InterruptCast(caster *model.Player) {
	val, ok := cm.activeCasts.LoadAndDelete(caster.ObjectID())
	if !ok {
		return
	}
	ac := val.(*activeCast)

	// Signal the finishCast goroutine to abort
	close(ac.cancel)
	caster.SetCasting(false)

	// Broadcast MagicSkillCanceled to all nearby players
	msc := serverpackets.NewMagicSkillCanceled(int32(caster.ObjectID()))
	mscData, err := msc.Write()
	if err != nil {
		slog.Error("failed to write MagicSkillCanceled", "error", err)
		return
	}
	cm.broadcastFunc(caster, mscData, len(mscData))
	cm.sendPacketFunc(caster.ObjectID(), mscData, len(mscData))

	slog.Debug("cast interrupted",
		"caster", caster.Name(),
		"skillID", ac.skillID)
}

// IsOnCooldown checks if a skill is on cooldown for a player.
func (cm *CastManager) IsOnCooldown(objectID uint32, skillID int32) bool {
	cdKey := cooldownKey(objectID, skillID)
	if expiry, ok := cm.cooldowns.Load(cdKey); ok {
		return time.Now().Before(expiry.(time.Time))
	}
	return false
}

// CooldownEntry represents a single skill cooldown for packet serialization.
type CooldownEntry struct {
	SkillID   int32
	ReuseMs   int32 // total reuse time in ms (from template)
	RemainMs  int32 // remaining time in ms
}

// GetAllCooldowns returns active cooldowns for a player.
// Used to build SkillCoolTime packet at login and on request.
func (cm *CastManager) GetAllCooldowns(objectID uint32) []CooldownEntry {
	now := time.Now()
	prefix := fmt.Sprintf("%d_", objectID)
	var result []CooldownEntry

	cm.cooldowns.Range(func(key, value any) bool {
		k := key.(string)
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			return true
		}
		expiry := value.(time.Time)
		remaining := expiry.Sub(now)
		if remaining <= 0 {
			cm.cooldowns.Delete(key)
			return true
		}

		// Parse skillID from key suffix
		var skillID int32
		if _, err := fmt.Sscanf(k[len(prefix):], "%d", &skillID); err != nil {
			return true
		}

		// Get reuse delay from skill template
		reuseMs := int32(remaining.Milliseconds())

		result = append(result, CooldownEntry{
			SkillID:  skillID,
			ReuseMs:  reuseMs, // approximation — we don't store original reuse
			RemainMs: reuseMs,
		})
		return true
	})

	return result
}

// checkRange validates that target is within CastRange of the caster.
// Uses squared distance to avoid sqrt (performance).
func (cm *CastManager) checkRange(caster *model.Player, targetObjID uint32, castRange int32) error {
	if cm.getWorldObject == nil {
		return nil // No resolver — skip range check
	}
	targetObj, ok := cm.getWorldObject(targetObjID)
	if !ok {
		return fmt.Errorf("target %d not found in world", targetObjID)
	}
	casterLoc := caster.Location()
	targetLoc := targetObj.Location()
	distSq := casterLoc.DistanceSquared(targetLoc)
	rangeSq := int64(castRange) * int64(castRange)
	if distSq > rangeSq {
		return fmt.Errorf("target out of range: dist²=%d, range²=%d", distSq, rangeSq)
	}
	return nil
}

// finishCast waits for HitTime then applies effects, with interrupt support.
// The ac.cancel channel is closed if the cast is interrupted.
func (cm *CastManager) finishCast(caster *model.Player, targetObjID uint32, tmpl *data.SkillTemplate, level int32, ac *activeCast) {
	timer := time.NewTimer(time.Duration(tmpl.HitTime) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ac.cancel:
		// Cast was interrupted — do nothing (InterruptCast already sent packets)
		return
	case <-timer.C:
		// Cast completed normally
	}

	// Clear casting state
	cm.activeCasts.Delete(caster.ObjectID())
	caster.SetCasting(false)

	// Check if caster is still alive
	if caster.IsDead() {
		return
	}

	// Revalidate target: check alive and in effect range
	if targetObjID != caster.ObjectID() {
		effectRange := tmpl.EffectRange
		if effectRange <= 0 {
			effectRange = tmpl.CastRange
		}
		if effectRange > 0 {
			if err := cm.checkRange(caster, targetObjID, effectRange); err != nil {
				slog.Debug("cast target out of range at finish",
					"caster", caster.Name(),
					"skill", tmpl.Name,
					"error", err)
				return
			}
		}
		// Check target is still alive (for offensive skills)
		if tmpl.IsDebuff && cm.getWorldObject != nil {
			if targetObj, ok := cm.getWorldObject(targetObjID); ok {
				if p, ok := targetObj.Data.(*model.Player); ok && p.IsDead() {
					slog.Debug("target died during cast",
						"caster", caster.Name(),
						"skill", tmpl.Name)
					return
				}
			}
		}
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

// isBeneficial returns true if the effect is beneficial (buff, heal, defensive).
func isBeneficial(e Effect) bool {
	switch e.Name() {
	case "Buff", "Heal", "MpHeal", "HealOverTime", "SpeedChange", "StatUp",
		"Reflect", "Transform", "Summon", "Cubic", "Resurrect", "Teleport":
		return true
	default:
		return false
	}
}

// cooldownKey generates a unique key for cooldown tracking.
func cooldownKey(objectID uint32, skillID int32) string {
	return fmt.Sprintf("%d_%d", objectID, skillID)
}

// UseItemSkill handles skill casting triggered by item use.
// Unlike UseMagic, this does NOT require the player to have learned the skill.
// The skill ID and level come from the item template, not the player's skill list.
//
// Phase 51: Item Handler System.
// Java reference: ItemSkillsTemplate.java → activeChar.useMagic(itemSkill, ...)
func (cm *CastManager) UseItemSkill(caster *model.Player, skillID, skillLevel int32) error {
	// 1. Get skill template
	tmpl := data.GetSkillTemplate(skillID, skillLevel)
	if tmpl == nil {
		return fmt.Errorf("item skill template not found: %d L%d", skillID, skillLevel)
	}

	// 2. Check if already casting
	if caster.IsCasting() {
		return fmt.Errorf("already casting")
	}

	// 3. Check cooldown (use item skill cooldown key)
	cdKey := cooldownKey(caster.ObjectID(), skillID)
	if expiry, ok := cm.cooldowns.Load(cdKey); ok {
		if time.Now().Before(expiry.(time.Time)) {
			return fmt.Errorf("item skill %d on cooldown", skillID)
		}
		cm.cooldowns.Delete(cdKey)
	}

	// 4. Check if dead
	if caster.IsDead() {
		return fmt.Errorf("cannot use item while dead")
	}

	// 5. Resolve target
	targetObjID := cm.resolveTarget(caster, tmpl)

	// 6. Broadcast MagicSkillUse (most item skills are instant: hitTime=0)
	loc := caster.Location()
	targetLoc := loc
	if target := caster.Target(); target != nil && target.ObjectID() == targetObjID {
		targetLoc = target.Location()
	}
	msu := serverpackets.NewMagicSkillUse(
		int32(caster.ObjectID()),
		int32(targetObjID),
		skillID,
		skillLevel,
		tmpl.HitTime,
		tmpl.ReuseDelay,
		loc.X, loc.Y, loc.Z,
		targetLoc.X, targetLoc.Y, targetLoc.Z,
		false,
	)
	msuData, err := msu.Write()
	if err != nil {
		return fmt.Errorf("writing MagicSkillUse for item: %w", err)
	}
	cm.broadcastFunc(caster, msuData, len(msuData))
	cm.sendPacketFunc(caster.ObjectID(), msuData, len(msuData))

	// 7. Apply effects (most item skills are instant)
	if tmpl.HitTime > 0 {
		caster.SetCasting(true)
		ac := &activeCast{
			cancel:  make(chan struct{}),
			skillID: skillID,
		}
		cm.activeCasts.Store(caster.ObjectID(), ac)
		go cm.finishCast(caster, targetObjID, tmpl, skillLevel, ac)
	} else {
		cm.applyEffects(caster, targetObjID, tmpl, skillLevel)
		cm.broadcastLaunched(caster, targetObjID, tmpl, skillLevel)
	}

	// 8. Start cooldown
	if tmpl.ReuseDelay > 0 {
		cm.cooldowns.Store(cdKey, time.Now().Add(time.Duration(tmpl.ReuseDelay)*time.Millisecond))
	}

	slog.Debug("item skill used",
		"caster", caster.Name(),
		"skill", tmpl.Name,
		"skillID", skillID,
		"level", skillLevel)

	return nil
}

// CastMgr is the global CastManager instance, initialized in main.go.
// Phase 5.9.4: Cast Flow & Packets.
var CastMgr *CastManager
