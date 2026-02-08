package gslistener

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/gameserver"
)

func TestNewGSConnection(t *testing.T) {
	// Create test connection
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Generate RSA key
	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	// Create GSConnection
	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Verify initial state
	assert.Equal(t, gameserver.GSStateConnected, conn.State())
	assert.NotNil(t, conn.BlowfishCipher())
	assert.Equal(t, rsaKey, conn.rsaKeyPair)
	assert.Nil(t, conn.GameServerInfo())
}

func TestGSConnectionStateMachine(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	// Initial: CONNECTED
	assert.Equal(t, gameserver.GSStateConnected, conn.State())

	// CONNECTED → BF_CONNECTED
	conn.SetState(gameserver.GSStateBFConnected)
	assert.Equal(t, gameserver.GSStateBFConnected, conn.State())

	// BF_CONNECTED → AUTHED
	conn.SetState(gameserver.GSStateAuthed)
	assert.Equal(t, gameserver.GSStateAuthed, conn.State())
}

func TestGSConnectionBlowfishSwap(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	// Initial cipher (22 bytes)
	cipher1 := conn.BlowfishCipher()
	require.NotNil(t, cipher1)

	// Swap to new cipher (40 bytes session key)
	sessionKey := make([]byte, 40)
	for i := range sessionKey {
		sessionKey[i] = byte(i)
	}
	cipher2, err := crypto.NewBlowfishCipher(sessionKey)
	require.NoError(t, err)

	conn.SetBlowfishCipher(cipher2)
	assert.Equal(t, cipher2, conn.BlowfishCipher())

	// Race test: concurrent reads
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := conn.BlowfishCipher()
			assert.NotNil(t, c)
		}()
	}
	wg.Wait()
}

func TestGSConnectionAccountTracking(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	// Add accounts
	conn.AddAccount("player1")
	conn.AddAccount("player2")

	assert.True(t, conn.HasAccount("player1"))
	assert.True(t, conn.HasAccount("player2"))
	assert.False(t, conn.HasAccount("player3"))

	// Remove account
	conn.RemoveAccount("player1")
	assert.False(t, conn.HasAccount("player1"))
	assert.True(t, conn.HasAccount("player2"))

	// Race test: concurrent add/remove/check
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		account := string(rune('A' + (i % 26)))
		go func() {
			defer wg.Done()
			conn.AddAccount(account)
			_ = conn.HasAccount(account)
			conn.RemoveAccount(account)
		}()
	}
	wg.Wait()
}

func TestGSConnectionAttachInfo(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	rsaKey, err := crypto.GenerateRSAKeyPair512()
	require.NoError(t, err)

	conn, err := NewGSConnection(server, rsaKey)
	require.NoError(t, err)

	// Initially nil
	assert.Nil(t, conn.GameServerInfo())

	// Attach info
	info := gameserver.NewGameServerInfo(1, []byte{0x01, 0x02, 0x03, 0x04})
	conn.AttachGameServerInfo(info)

	assert.NotNil(t, conn.GameServerInfo())
	assert.Equal(t, 1, conn.GameServerInfo().ID())
}
