package quests

import (
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/quest"
)

// RegisterAllQuests creates and registers all implemented quests into the manager.
func RegisterAllQuests(m *quest.Manager) error {
	constructors := []func() *quest.Quest{
		NewQ00001, // Letters of Love
		NewQ00002, // What Women Want
		NewQ00003, // Will The Seal Be Broken
		NewQ00004, // Long Live the Pa'agrio Lord
		NewQ00005, // Miner's Favor
		NewQ00101, // Sword of Solidarity
		NewQ00102, // Sea of Spores Fever
		NewQ00151, // Cure for Fever Disease
		NewQ00152, // Shards of Golem
		NewQ00257, // The Guard is Busy
		NewQ00260, // Hunt the Orcs
		NewQ00261, // Collector's Dream
		NewQ00293, // The Hidden Veins
		NewQ00320, // Bones Tell the Future
	}

	for _, ctor := range constructors {
		q := ctor()
		if err := m.RegisterQuest(q); err != nil {
			return fmt.Errorf("register quest %q (ID=%d): %w", q.Name(), q.ID(), err)
		}
	}

	slog.Info("quests registered", "count", len(constructors))
	return nil
}
