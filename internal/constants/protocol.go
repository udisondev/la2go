package constants

// L2 Interlude Protocol Constants
//
// This file contains all protocol-level constants for Lineage 2 Chronicle: Interlude (C6).
// These values are defined by the original L2J server implementation and L2 client protocol.

// Protocol Revision Constants
const (
	// ProtocolRevisionInterlude is the GS↔LS protocol revision for Interlude
	ProtocolRevisionInterlude = 0x0106

	// ProtocolRevisionInit is the protocol revision field in Init packet (Client↔LS)
	// Format differs from GS↔LS revision (0x0000C621 vs 0x0106)
	ProtocolRevisionInit = 0x0000C621
)

// RSA Key Size Constants
const (
	// RSAKeyBits is the RSA key size in bits for Client↔LoginServer (1024-bit)
	RSAKeyBits = 1024

	// RSA512KeyBits is the RSA key size in bits for GameServer↔LoginServer (512-bit)
	RSA512KeyBits = 512

	// RSAPublicExponent is the RSA public exponent (F4 = 65537)
	RSAPublicExponent = 65537

	// RSA1024ModulusSize is the RSA-1024 modulus size in bytes (1024 bits / 8)
	RSA1024ModulusSize = 128

	// RSA512ModulusSize is the RSA-512 modulus size in bytes (512 bits / 8)
	RSA512ModulusSize = 64

	// RSAModulusMaxSize is the maximum modulus size with potential leading zero byte
	// Java BigInteger.toByteArray() may return N+1 bytes with leading 0x00
	RSAModulusMaxSize = RSA1024ModulusSize + 1 // 129 bytes
)

// Blowfish Cipher Constants
const (
	// BlowfishKeySize is the Blowfish key size in bytes (128-bit)
	BlowfishKeySize = 16

	// BlowfishBlockSize is the Blowfish block size in bytes (64-bit)
	BlowfishBlockSize = 8
)

// Packet Structure Constants
const (
	// PacketHeaderSize is the packet length header size (2 bytes, little-endian uint16)
	PacketHeaderSize = 2

	// PacketChecksumSize is the XOR checksum size in bytes (32-bit)
	PacketChecksumSize = 4

	// PacketPaddingAlign is the padding alignment for encrypted packets (Blowfish requires 8-byte blocks)
	PacketPaddingAlign = 8

	// PacketBufferPadding is the extra buffer space for encryption overhead
	PacketBufferPadding = 16
)

// Init Packet Structure Constants (Client↔LoginServer)
//
// Init packet format (opcode 0x00):
//   [opcode 1 byte]
//   [sessionID 4 bytes LE]
//   [protocolRevision 4 bytes LE]
//   [scrambled RSA modulus 128 bytes]
//   [GameGuard constants 16 bytes (4×uint32 LE)]
//   [Blowfish key 16 bytes]
//   [null terminator 1 byte]
//   Total: 170 bytes
const (
	// InitPacketOpcodeOffset is the offset of opcode field
	InitPacketOpcodeOffset = 0

	// InitPacketSessionIDOffset is the offset of sessionID field (4 bytes LE)
	InitPacketSessionIDOffset = 1

	// InitPacketProtocolRevOffset is the offset of protocol revision field (4 bytes LE)
	InitPacketProtocolRevOffset = 5

	// InitPacketModulusOffset is the offset of scrambled RSA modulus (128 bytes)
	InitPacketModulusOffset = 9

	// InitPacketGGConstantsOffset is the offset of GameGuard constants (4×4 bytes LE)
	InitPacketGGConstantsOffset = 137

	// InitPacketBlowfishKeyOffset is the offset of Blowfish key (16 bytes)
	InitPacketBlowfishKeyOffset = 153

	// InitPacketNullTerminatorOffset is the offset of null terminator (1 byte)
	InitPacketNullTerminatorOffset = 169

	// InitPacketTotalSize is the total size of Init packet plaintext
	InitPacketTotalSize = 170
)

// GameGuard Constants (Anti-cheat handshake)
//
// These are protocol-defined constants sent in Init packet for GameGuard verification.
// Note: GGConst3 uses negative literal to avoid int32 overflow (0x97ADB620 > MaxInt32).
const (
	GGConst1 = 0x29DD954E
	GGConst2 = 0x77C39CFC
	GGConst3 = 0x97ADB620 // -0x685249E0 as uint32 (overflow fix from MEMORY.md)
	GGConst4 = 0x07BDE0F7
)

// RequestAuthLogin Packet Structure Constants
//
// After RSA decryption, the RequestAuthLogin packet contains username and password
// at fixed offsets within the 128-byte decrypted block.
const (
	// AuthLoginUsernameOffset is the offset of username field in RSA-decrypted data
	AuthLoginUsernameOffset = 0x5E // 94 bytes

	// AuthLoginUsernameMaxLength is the maximum username length in bytes
	AuthLoginUsernameMaxLength = 14

	// AuthLoginPasswordOffset is the offset of password field in RSA-decrypted data
	AuthLoginPasswordOffset = 0x6C // 108 bytes

	// AuthLoginPasswordMaxLength is the maximum password length in bytes
	AuthLoginPasswordMaxLength = 16
)

// LoginOk Packet Constants
const (
	// LoginOkUnknownField is an unknown constant field in LoginOk packet at offset 17
	// Value observed in L2J implementation, purpose unclear
	LoginOkUnknownField = 0x000003EA

	// LoginOkPaddingSize is the padding size at the end of LoginOk packet
	LoginOkPaddingSize = 16
)

// RSA Scrambling Algorithm Constants
//
// The scrambling algorithm obfuscates the RSA modulus using a 4-step XOR/swap process.
// These offsets are defined by the L2J ScrambledKeyPair implementation.
const (
	// ScrambleSwapOffset1 is the first swap range start offset
	ScrambleSwapOffset1 = 0x00

	// ScrambleSwapOffset2 is the second swap range start offset
	ScrambleSwapOffset2 = 0x4D // 77 bytes

	// ScrambleSwapLength is the number of bytes to swap
	ScrambleSwapLength = 4

	// ScrambleXORBlock1Start is the first XOR block start offset
	ScrambleXORBlock1Start = 0x00

	// ScrambleXORBlock1Size is the first XOR block size
	ScrambleXORBlock1Size = 0x40 // 64 bytes

	// ScrambleXORBlock2Start is the second XOR block start offset
	ScrambleXORBlock2Start = 0x40 // 64 bytes

	// ScrambleXOROffset1 is the first specific XOR offset
	ScrambleXOROffset1 = 0x0D // 13 bytes

	// ScrambleXOROffset2 is the second specific XOR offset
	ScrambleXOROffset2 = 0x34 // 52 bytes

	// ScrambleXORLength is the number of bytes to XOR at specific offsets
	ScrambleXORLength = 4
)

// XOR Encryption Constants (Init Packet Pre-encryption)
//
// The Init packet uses a custom XOR encryption pass before Blowfish encryption.
// Algorithm matches L2J NewCrypt.encXORPass implementation.
const (
	// XOREncryptSkipBytes is the number of bytes to skip at start (sessionID not encrypted)
	XOREncryptSkipBytes = 4

	// XOREncryptStopOffset is the offset from end where XOR encryption stops
	XOREncryptStopOffset = 8
)

// Buffer Pool Size Constants
const (
	// DefaultSendBufSize is the default send buffer size for LoginServer client connections
	DefaultSendBufSize = 512

	// DefaultReadBufSize is the default read buffer size for LoginServer client connections
	DefaultReadBufSize = 512

	// GSListenerSendBufSize is the send buffer size for GameServer↔LoginServer connections
	GSListenerSendBufSize = 1024

	// GSListenerReadBufSize is the read buffer size for GameServer↔LoginServer connections
	GSListenerReadBufSize = 8192
)

// ObjectID Range Constants (L2J Mobius Interlude protocol)
// These ranges are used to distinguish object types by their ID.
const (
	// ObjectIDPlayerStart is the start of Player object ID range (0x10000000 = 268435456)
	ObjectIDPlayerStart = 0x10000000

	// ObjectIDPlayerEnd is the end of Player object ID range (0x1FFFFFFF = 536870911)
	ObjectIDPlayerEnd = 0x1FFFFFFF

	// ObjectIDNpcStart is the start of NPC object ID range (0x20000000 = 536870912)
	ObjectIDNpcStart = 0x20000000

	// ObjectIDItemStart is the start of Item (on ground) object ID range (0x00000001 = 1)
	ObjectIDItemStart = 0x00000001

	// ObjectIDItemEnd is the end of Item object ID range (0x0FFFFFFF = 268435455)
	ObjectIDItemEnd = 0x0FFFFFFF
)

// Server Default Constants
const (
	// DefaultMaxPlayers is the default maximum players per game server
	DefaultMaxPlayers = 1000

	// DefaultServerType is the default server type (1 = Normal)
	DefaultServerType = 1

	// DefaultServerStatus is the default server status (1 = Online, 0 = Down)
	DefaultServerStatus = 1
)
