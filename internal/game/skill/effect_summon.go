package skill

import (
	"log/slog"
	"strconv"
)

// SummonEffect summons a servitor/pet NPC bound to the caster.
// Params: "npcID" (int32 template ID), "summonType" (string: "SERVITOR", "PET").
//
// Java reference: Summon.java — spawns servitor with caster as owner.
type SummonEffect struct {
	npcID      int32
	summonType string
}

func NewSummonEffect(params map[string]string) Effect {
	npcID, _ := strconv.Atoi(params["npcID"])
	summonType := params["summonType"]
	if summonType == "" {
		summonType = "SERVITOR"
	}
	return &SummonEffect{
		npcID:      int32(npcID),
		summonType: summonType,
	}
}

func (e *SummonEffect) Name() string    { return "Summon" }
func (e *SummonEffect) IsInstant() bool { return true }

func (e *SummonEffect) OnStart(casterObjID, _ uint32) {
	slog.Debug("summon",
		"npcID", e.npcID,
		"type", e.summonType,
		"owner", casterObjID)
	// Actual spawning handled by CastManager (needs World + SpawnManager access)
}

func (e *SummonEffect) OnActionTime(_, _ uint32) bool { return false }
func (e *SummonEffect) OnExit(_, _ uint32)            {}

// NpcID returns the NPC template ID to summon.
func (e *SummonEffect) NpcID() int32 { return e.npcID }

// SummonType returns the summon type (SERVITOR or PET).
func (e *SummonEffect) SummonType() string { return e.summonType }
