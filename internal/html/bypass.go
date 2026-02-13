package html

import (
	"fmt"
	"strconv"
	"strings"
)

// BypassCommand represents a parsed NPC bypass command.
// Format: "npc_<objectID>_<command> [args...]"
type BypassCommand struct {
	ObjectID uint32
	Command  string   // "Shop", "Chat", "Sell", "Teleport", etc.
	Args     []string // arguments after space in command part
}

// allowedCommands is a whitelist of valid NPC bypass commands.
var allowedCommands = map[string]bool{
	"Shop":       true,
	"Sell":       true,
	"Chat":       true,
	"Teleport":   true,
	"Quest":      true,
	"DepositP":   true,
	"WithdrawP":  true,
	"Multisell":  true,
	"Link":       true,
	"Craft":      true,
	"EnchantSkillList": true,
}

// ParseNpcBypass parses a bypass string in format "npc_<objectID>_<command> [args...]".
func ParseNpcBypass(bypass string) (*BypassCommand, error) {
	if !strings.HasPrefix(bypass, "npc_") {
		return nil, fmt.Errorf("not an NPC bypass: %s", bypass)
	}

	// Split "npc_<objectID>_<command> args" into ["npc", "<objectID>", "<command> args"]
	parts := strings.SplitN(bypass, "_", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed NPC bypass: %s", bypass)
	}

	objectID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid objectID in bypass %q: %w", bypass, err)
	}

	// parts[2] may contain "Chat 1" or just "Shop"
	commandPart := parts[2]
	cmdParts := strings.SplitN(commandPart, " ", 2)
	cmdName := cmdParts[0]

	if !allowedCommands[cmdName] {
		return nil, fmt.Errorf("unknown bypass command: %s", cmdName)
	}

	var args []string
	if len(cmdParts) > 1 && cmdParts[1] != "" {
		args = strings.Fields(cmdParts[1])
	}

	return &BypassCommand{
		ObjectID: uint32(objectID),
		Command:  cmdName,
		Args:     args,
	}, nil
}
