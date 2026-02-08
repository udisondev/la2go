package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gslistener/packet"
)

// ServerStatus [0x06] — GS → LS обновление статуса сервера
//
// Format:
//   [opcode 0x06]
//   [count int32]
//   for each count:
//     [attributeID int32]
//     [value int32]
//
// Attribute IDs:
//   0 = showingBrackets (value: 0/1)
//   1 = serverType
//   2 = status
//   3 = ageLimit
//   4 = maxPlayers
type ServerStatus struct {
	Attributes []Attribute
}

// Attribute представляет пару (id, value) в пакете ServerStatus.
type Attribute struct {
	ID    int32
	Value int32
}

// Parse парсит пакет ServerStatus из body (без opcode).
func (p *ServerStatus) Parse(body []byte) error {
	r := packet.NewReader(body)

	// Читаем count
	count, err := r.ReadInt()
	if err != nil {
		return fmt.Errorf("reading count: %w", err)
	}

	if count < 0 || count > 100 {
		return fmt.Errorf("invalid count: %d", count)
	}

	// Читаем атрибуты
	p.Attributes = make([]Attribute, 0, count)
	for range count {
		attrID, err := r.ReadInt()
		if err != nil {
			return fmt.Errorf("reading attribute ID: %w", err)
		}

		value, err := r.ReadInt()
		if err != nil {
			return fmt.Errorf("reading attribute value: %w", err)
		}

		p.Attributes = append(p.Attributes, Attribute{
			ID:    attrID,
			Value: value,
		})
	}

	return nil
}
