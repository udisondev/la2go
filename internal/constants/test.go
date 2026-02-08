package constants

import "time"

// Test Constants
//
// IMPORTANT: These constants are for testing only. DO NOT use in production code.

// Integration Test Timeout Constants
const (
	// TestServerStartupDelay is the delay to wait for server startup in integration tests
	TestServerStartupDelay = 100 * time.Millisecond

	// TestGracefulShutdownWait is the delay to wait for graceful shutdown in tests
	TestGracefulShutdownWait = 100 * time.Millisecond
)

// Concurrency Test Constants
const (
	// TestConcurrentClientsSmall is the number of concurrent clients for small load tests
	TestConcurrentClientsSmall = 10

	// TestConcurrentClientsLarge is the number of concurrent clients for large load tests
	TestConcurrentClientsLarge = 20
)

// Test Server Configuration Constants
const (
	// TestMaxPlayers is the max players value used in test fixtures
	TestMaxPlayers = 1000

	// TestServerPort is the default GameServer port used in tests (8180 = 0x1FF4)
	TestServerPort = 8180

	// TestServerPortLE1 is the low byte of TestServerPort in little-endian (0xF4)
	TestServerPortLE1 = 0xF4

	// TestServerPortLE2 is the high byte of TestServerPort in little-endian (0x1F)
	TestServerPortLE2 = 0x1F

	// TestMaxPlayersLE1 is the low byte of TestMaxPlayers in little-endian (0xE8)
	TestMaxPlayersLE1 = 0xE8

	// TestMaxPlayersLE2 is the high byte of TestMaxPlayers in little-endian (0x03)
	TestMaxPlayersLE2 = 0x03
)

// Test Packet Size Constants
const (
	// TestInitPacketBufSize is the buffer size for testing Init packet (256 bytes > 170 required)
	TestInitPacketBufSize = 256
)
