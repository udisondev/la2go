// Package crest provides clan and alliance crest management
// for the Lineage 2 crest system.
package crest

import "errors"

// CrestType identifies the kind of crest.
type CrestType int32

const (
	// Pledge is a standard clan crest (max 256 bytes, 16x12 BMP).
	Pledge CrestType = 1
	// PledgeLarge is a large clan crest (max 2176 bytes, 64x32 BMP).
	PledgeLarge CrestType = 2
	// Ally is an alliance crest (max 192 bytes, 8x12 BMP).
	Ally CrestType = 3
)

// Size limits for each crest type (bytes).
const (
	MaxPledgeSize      = 256
	MaxPledgeLargeSize = 2176
	MaxAllySize        = 192
)

// Common errors.
var (
	ErrDataTooLarge = errors.New("crest data exceeds maximum size")
	ErrInvalidType  = errors.New("invalid crest type")
)

// maxSize returns the byte limit for a given crest type.
func maxSize(t CrestType) (int, error) {
	switch t {
	case Pledge:
		return MaxPledgeSize, nil
	case PledgeLarge:
		return MaxPledgeLargeSize, nil
	case Ally:
		return MaxAllySize, nil
	default:
		return 0, ErrInvalidType
	}
}

// Crest holds the crest image data.
type Crest struct {
	id   int32
	data []byte
	typ  CrestType
}

// ID returns the crest identifier.
func (c *Crest) ID() int32 { return c.id }

// Data returns the raw crest image bytes.
func (c *Crest) Data() []byte { return c.data }

// Type returns the crest type.
func (c *Crest) Type() CrestType { return c.typ }
