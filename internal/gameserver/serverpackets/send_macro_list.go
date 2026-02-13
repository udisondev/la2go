package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeSendMacroList is the S2C opcode 0xE7.
const OpcodeSendMacroList byte = 0xE7

// SendMacroList sends macro data to the client.
type SendMacroList struct {
	Revision int32        // revision counter (changes on every edit)
	Count    int8         // total macro count for the player
	Macro    *model.Macro // macro to send (nil = empty update)
}

// Write serializes the packet.
func (p *SendMacroList) Write() ([]byte, error) {
	w := packet.NewWriter(256)
	w.WriteByte(OpcodeSendMacroList)
	w.WriteInt(p.Revision)
	w.WriteByte(0) // unknown

	w.WriteByte(byte(p.Count))

	hasMacro := byte(0)
	if p.Macro != nil {
		hasMacro = 1
	}
	w.WriteByte(hasMacro)

	if p.Macro != nil {
		w.WriteInt(p.Macro.ID)
		w.WriteString(p.Macro.Name)
		w.WriteString(p.Macro.Desc)
		w.WriteString(p.Macro.Acronym)
		w.WriteByte(byte(p.Macro.Icon))
		w.WriteByte(byte(len(p.Macro.Commands)))

		for i, cmd := range p.Macro.Commands {
			w.WriteByte(byte(i))
			w.WriteByte(byte(cmd.Type))
			w.WriteInt(cmd.D1)
			w.WriteByte(byte(cmd.D2))
			w.WriteString(cmd.Command)
		}
	}

	return w.Bytes(), nil
}
