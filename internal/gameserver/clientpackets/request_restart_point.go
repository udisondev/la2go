package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestRestartPoint is the C2S opcode 0x6D.
// Client sends this after death to select respawn location.
const OpcodeRequestRestartPoint byte = 0x6D

// Restart point types (from Java RequestRestartPoint.java).
const (
	RestartPointVillage  int32 = 0 // default — nearest village
	RestartPointClanHall int32 = 1 // clan hall
	RestartPointCastle   int32 = 2 // castle
	RestartPointSiegeHQ  int32 = 3 // siege headquarters
	RestartPointFixed    int32 = 4 // fixed res (scroll)
)

// RequestRestartPoint represents a respawn location choice.
type RequestRestartPoint struct {
	PointType int32 // 0=village, 1=clanhall, 2=castle, 3=siegeHQ, 4=fixed
}

// ParseRequestRestartPoint parses the packet from raw bytes.
func ParseRequestRestartPoint(data []byte) (*RequestRestartPoint, error) {
	r := packet.NewReader(data)

	pointType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading PointType: %w", err)
	}

	return &RequestRestartPoint{
		PointType: pointType,
	}, nil
}
