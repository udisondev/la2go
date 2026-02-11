package skill

import "fmt"

// effectRegistry maps effect name → factory function.
// Populated by init() functions in individual effect files.
//
// Phase 5.9.3: Effect Framework.
// Java reference: EffectHandler.java
var effectRegistry = map[string]func(params map[string]string) Effect{}

// RegisterEffect registers an effect factory by name.
// Called from init() in each effect implementation file.
func RegisterEffect(name string, factory func(params map[string]string) Effect) {
	effectRegistry[name] = factory
}

// CreateEffect creates an effect by name using the registered factory.
// Returns error if name is not registered.
func CreateEffect(name string, params map[string]string) (Effect, error) {
	factory, ok := effectRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown effect type: %s", name)
	}
	return factory(params), nil
}

func init() {
	RegisterEffect("Buff", NewBuffEffect)
	RegisterEffect("PhysicalDamage", NewPhysicalDamageEffect)
	RegisterEffect("MagicalDamage", NewMagicalDamageEffect)
	RegisterEffect("Heal", NewHealEffect)
	RegisterEffect("MpHeal", NewMpHealEffect)
	RegisterEffect("HpDrain", NewHpDrainEffect)
	RegisterEffect("DamageOverTime", NewDamageOverTimeEffect)
	RegisterEffect("HealOverTime", NewHealOverTimeEffect)
	RegisterEffect("Stun", NewStunEffect)
	RegisterEffect("Root", NewRootEffect)
	RegisterEffect("Paralyze", NewParalyzeEffect)
	RegisterEffect("Sleep", NewSleepEffect)
	RegisterEffect("CancelTarget", NewCancelTargetEffect)
	RegisterEffect("SpeedChange", NewSpeedChangeEffect)
	RegisterEffect("StatUp", NewStatUpEffect)
}
