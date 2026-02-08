package testutil

import (
	"github.com/udisondev/la2go/internal/crypto"
	"github.com/udisondev/la2go/internal/login"
)

// Fixtures содержит предварительно сгенерированные тестовые данные
// для избежания дублирования в тестах.
var Fixtures = struct {
	// RSA ключи (генерируются один раз при init)
	RSAKey    *crypto.RSAKeyPair
	RSAKey512 *crypto.RSAKeyPair

	// Blowfish ключ (16 байт)
	BlowfishKey []byte

	// SessionKey для тестов
	SessionKey login.SessionKey

	// Тестовые аккаунты
	ValidAccount  string
	ValidPassword string
	ValidHash     string // SHA-1 hash от ValidPassword

	// Game Server test data
	GSServerID   byte
	GSBlowfishKey []byte
	GSHexID      string
}{
	BlowfishKey: []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
	},
	SessionKey: login.SessionKey{
		LoginOkID1: 12345,
		LoginOkID2: 67890,
		PlayOkID1:  11111,
		PlayOkID2:  22222,
	},
	ValidAccount:  "testuser",
	ValidPassword: "testpass",
	// SHA-1("testpass") в hex
	ValidHash: "206c80413b9a96c1312cc346b7d2517b84463edd",

	GSServerID: 1,
	GSBlowfishKey: []byte{
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20,
	},
	GSHexID: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
}

func init() {
	var err error

	// Генерируем RSA-2048 ключ
	Fixtures.RSAKey, err = crypto.GenerateRSAKeyPair()
	if err != nil {
		panic("failed to generate RSA-2048 key: " + err.Error())
	}

	// Генерируем RSA-512 ключ
	Fixtures.RSAKey512, err = crypto.GenerateRSAKeyPair512()
	if err != nil {
		panic("failed to generate RSA-512 key: " + err.Error())
	}
}
