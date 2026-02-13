package crest

import "context"

// CrestRow is the DB transfer object for a crest.
type CrestRow struct {
	CrestID int32
	Data    []byte
	Type    int32
}

// Store abstracts crest persistence.
// Declared in the consuming package per Go interface conventions.
type Store interface {
	LoadAll(ctx context.Context) ([]CrestRow, error)
	Insert(ctx context.Context, row CrestRow) error
	Delete(ctx context.Context, crestID int32) error
}
