package serverpackets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/gameserver"
)

func TestLoginServerFail(t *testing.T) {
	buf := make([]byte, 256)

	tests := []struct {
		name   string
		reason byte
	}{
		{"IP banned", gameserver.ReasonIPBanned},
		{"wrong hex ID", gameserver.ReasonWrongHexID},
		{"not authed", gameserver.ReasonNotAuthed},
		{"already logged in", gameserver.ReasonAlreadyLoggedIn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := LoginServerFail(buf, tt.reason)

			// Verify size: opcode(1) + reason(1) = 2
			require.Equal(t, 2, n)

			// Verify opcode
			assert.Equal(t, byte(0x01), buf[0])

			// Verify reason
			assert.Equal(t, tt.reason, buf[1])
		})
	}
}
