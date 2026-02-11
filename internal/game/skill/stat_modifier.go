package skill

// StatModType defines how a stat modifier is applied.
//
// Phase 5.9.3: Effect Framework.
// Java reference: Stat.java, FuncAdd.java, FuncMul.java
type StatModType int8

const (
	StatModAdd StatModType = iota // Additive bonus (e.g. +100 pAtk)
	StatModMul                    // Multiplicative bonus (e.g. ×1.2 speed)
)

// StatModifier represents a single stat modification from an effect.
// Multiple modifiers can stack on the same stat.
//
// Phase 5.9.3: Effect Framework.
type StatModifier struct {
	Stat  string      // "pAtk", "pDef", "speed", "mAtk", "mDef", "maxHP", "maxMP"
	Type  StatModType // ADD or MUL
	Value float64     // bonus value
}
