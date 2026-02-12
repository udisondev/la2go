package combat

import (
	"time"

	"github.com/udisondev/la2go/internal/model"
)

// AttackUntilHit выполняет атаки до первого попадания (non-miss, non-crit).
// Возвращает HitResult и число попыток.
func AttackUntilHit(mgr *CombatManager, attacker *model.Player, target *model.WorldObject, maxAttempts int) (HitResult, int) {
	resultCh := make(chan HitResult, 1)
	mgr.SetHitObserver(func(r HitResult) {
		select {
		case resultCh <- r:
		default:
		}
	})
	defer mgr.SetHitObserver(nil)

	for attempt := range maxAttempts {
		mgr.ExecuteAttack(attacker, target)
		result := <-resultCh
		time.Sleep(2 * time.Second) // продвигает fake clock → damage applies
		if !result.Miss && !result.Crit {
			return result, attempt + 1
		}
	}
	return HitResult{Miss: true}, maxAttempts
}

// AttackUntilDead выполняет атаки до смерти цели.
// Возвращает число попыток.
func AttackUntilDead(mgr *CombatManager, attacker *model.Player, target *model.WorldObject, targetChar *model.Character, maxAttempts int) int {
	resultCh := make(chan HitResult, 1)
	mgr.SetHitObserver(func(r HitResult) {
		select {
		case resultCh <- r:
		default:
		}
	})
	defer mgr.SetHitObserver(nil)

	for attempt := range maxAttempts {
		if targetChar.IsDead() {
			return attempt
		}
		mgr.ExecuteAttack(attacker, target)
		<-resultCh
		time.Sleep(2 * time.Second)
	}
	return maxAttempts
}

// NpcAttackUntilHit выполняет NPC атаки до первого попадания (non-miss, non-crit).
// Возвращает HitResult и число попыток.
func NpcAttackUntilHit(mgr *CombatManager, npc *model.Npc, target *model.WorldObject, maxAttempts int) (HitResult, int) {
	resultCh := make(chan HitResult, 1)
	mgr.SetHitObserver(func(r HitResult) {
		select {
		case resultCh <- r:
		default:
		}
	})
	defer mgr.SetHitObserver(nil)

	for attempt := range maxAttempts {
		mgr.ExecuteNpcAttack(npc, target)
		result := <-resultCh
		time.Sleep(3 * time.Second) // NPC attack delay
		if !result.Miss && !result.Crit {
			return result, attempt + 1
		}
	}
	return HitResult{Miss: true}, maxAttempts
}
