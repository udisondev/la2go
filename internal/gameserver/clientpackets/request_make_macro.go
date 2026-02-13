package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

// OpcodeRequestMakeMacro is the C2S opcode 0xC1.
const OpcodeRequestMakeMacro byte = 0xC1

// RequestMakeMacro represents a request to create or update a macro.
type RequestMakeMacro struct {
	Macro *model.Macro
}

// ParseRequestMakeMacro parses the packet from raw bytes.
func ParseRequestMakeMacro(data []byte) (*RequestMakeMacro, error) {
	r := packet.NewReader(data)

	id, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading macro ID: %w", err)
	}

	name, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading macro name: %w", err)
	}

	desc, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading macro desc: %w", err)
	}

	acronym, err := r.ReadString()
	if err != nil {
		return nil, fmt.Errorf("reading macro acronym: %w", err)
	}

	icon, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading macro icon: %w", err)
	}

	count, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading command count: %w", err)
	}

	if int(count) > model.MaxMacroCommands {
		return nil, fmt.Errorf("too many macro commands: %d (max %d)", count, model.MaxMacroCommands)
	}

	commands := make([]model.MacroCmd, 0, count)
	for range int(count) {
		entry, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("reading command entry: %w", err)
		}

		cmdType, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("reading command type: %w", err)
		}

		d1, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading command d1: %w", err)
		}

		d2, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("reading command d2: %w", err)
		}

		cmd, err := r.ReadString()
		if err != nil {
			return nil, fmt.Errorf("reading command string: %w", err)
		}

		commands = append(commands, model.MacroCmd{
			Entry:   int8(entry),
			Type:    model.MacroCmdType(cmdType),
			D1:      d1,
			D2:      int8(d2),
			Command: cmd,
		})
	}

	return &RequestMakeMacro{
		Macro: &model.Macro{
			ID:       id,
			Name:     name,
			Desc:     desc,
			Acronym:  acronym,
			Icon:     int8(icon),
			Commands: commands,
		},
	}, nil
}
