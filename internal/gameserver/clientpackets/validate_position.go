package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeValidatePosition is the client packet opcode for ValidatePosition.
// Client sends this periodically (~200ms) to report current position.
const OpcodeValidatePosition = 0x48

// ValidatePosition represents the client's position update packet.
// The client sends this packet periodically to keep the server informed
// of its current position, enabling desync detection.
//
// Packet structure:
//   - X (int32): Client-reported X coordinate
//   - Y (int32): Client-reported Y coordinate
//   - Z (int32): Client-reported Z coordinate
//   - Heading (int32): Client-reported heading (0-65535 as int32)
//   - VehicleID (int32): Vehicle objectID (0 if not in vehicle)
//
// Reference: ValidatePosition.java (L2J Mobius)
type ValidatePosition struct {
	X         int32 // Client-reported X coordinate
	Y         int32 // Client-reported Y coordinate
	Z         int32 // Client-reported Z coordinate
	Heading   int32 // Client-reported heading (stored as int32, not uint16)
	VehicleID int32 // Vehicle objectID (0 if not mounted)
}

// ParseValidatePosition parses a ValidatePosition packet from raw bytes.
//
// The packet format is:
//   - opcode (byte): 0x48
//   - x (int32): X coordinate
//   - y (int32): Y coordinate
//   - z (int32): Z coordinate
//   - heading (int32): heading value
//   - vehicleID (int32): vehicle objectID
//
// Returns an error if parsing fails.
func ParseValidatePosition(data []byte) (*ValidatePosition, error) {
	r := packet.NewReader(data)

	// Read coordinates (opcode already stripped by HandlePacket)
	x, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading X: %w", err)
	}

	y, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Y: %w", err)
	}

	z, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Z: %w", err)
	}

	// Read heading (Java reads as int32, not short)
	heading, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Heading: %w", err)
	}

	// Read vehicle ID
	vehicleID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading VehicleID: %w", err)
	}

	return &ValidatePosition{
		X:         x,
		Y:         y,
		Z:         z,
		Heading:   heading,
		VehicleID: vehicleID,
	}, nil
}
