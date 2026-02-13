package crest

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

// Table manages all loaded crests (analogous to Java CrestTable).
// Thread-safe: crests stored in sync.Map, ID generation via atomic.
type Table struct {
	crests sync.Map     // map[int32]*Crest
	nextID atomic.Int32 // monotonically increasing crest ID
}

// NewTable creates an empty crest table.
func NewTable() *Table {
	return &Table{}
}

// Init loads crests from DB rows and seeds the nextID counter.
// Must be called once at startup before any Create/Remove calls.
func (t *Table) Init(rows []CrestRow) {
	var maxID int32
	for _, row := range rows {
		ct := CrestType(row.Type)
		c := &Crest{
			id:   row.CrestID,
			data: row.Data,
			typ:  ct,
		}
		t.crests.Store(row.CrestID, c)

		if row.CrestID > maxID {
			maxID = row.CrestID
		}
	}
	// Следующий ID — на единицу больше максимального из БД.
	t.nextID.Store(maxID + 1)

	slog.Info("crest table initialized", "count", len(rows), "nextID", maxID+1)
}

// Crest returns a crest by ID, or nil if not found.
func (t *Table) Crest(id int32) *Crest {
	v, ok := t.crests.Load(id)
	if !ok {
		return nil
	}
	return v.(*Crest)
}

// CreateCrest validates size, assigns an atomic ID, and stores the crest.
// Returns the created crest or an error if data exceeds the type limit.
func (t *Table) CreateCrest(data []byte, typ CrestType) (*Crest, error) {
	limit, err := maxSize(typ)
	if err != nil {
		return nil, fmt.Errorf("create crest: %w", err)
	}
	if len(data) > limit {
		return nil, fmt.Errorf("create crest type %d: %w (got %d, max %d)",
			typ, ErrDataTooLarge, len(data), limit)
	}

	id := t.nextID.Add(1) - 1 // fetch-then-increment semantics

	c := &Crest{
		id:   id,
		data: data,
		typ:  typ,
	}
	t.crests.Store(id, c)
	return c, nil
}

// RemoveCrest deletes a crest from the table.
// The ID is never reused.
func (t *Table) RemoveCrest(id int32) {
	t.crests.Delete(id)
}
