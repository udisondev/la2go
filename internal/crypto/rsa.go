package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
)

// RSAKeyPair holds an RSA-1024 key pair and the scrambled modulus for the client.
type RSAKeyPair struct {
	PrivateKey       *rsa.PrivateKey
	ScrambledModulus []byte // 128 bytes, scrambled for L2 client
}

// GenerateRSAKeyPair generates an RSA-1024 key pair with exponent 65537 (F4)
// and pre-computes the scrambled modulus.
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key: %w", err)
	}

	modBytes := privateKey.PublicKey.N.Bytes()

	// Java BigInteger.toByteArray() may return 129 bytes with leading zero.
	// We need exactly 128 bytes.
	if len(modBytes) == 129 && modBytes[0] == 0 {
		modBytes = modBytes[1:]
	}
	if len(modBytes) < 128 {
		padded := make([]byte, 128)
		copy(padded[128-len(modBytes):], modBytes)
		modBytes = padded
	}

	scrambled := ScrambleModulus(modBytes)

	return &RSAKeyPair{
		PrivateKey:       privateKey,
		ScrambledModulus: scrambled,
	}, nil
}

// ScrambleModulus applies the 4-step XOR/swap obfuscation to the RSA modulus
// as done in L2J ScrambledKeyPair.java.
// Input must be exactly 128 bytes.
func ScrambleModulus(modulus []byte) []byte {
	if len(modulus) != 128 {
		panic(fmt.Sprintf("ScrambleModulus: expected 128 bytes, got %d", len(modulus)))
	}

	scrambled := make([]byte, 128)
	copy(scrambled, modulus)

	// Step 1: swap bytes 0x00-0x03 with 0x4D-0x50
	for i := range 4 {
		scrambled[i], scrambled[0x4D+i] = scrambled[0x4D+i], scrambled[i]
	}

	// Step 2: XOR first 0x40 bytes with last 0x40 bytes
	for i := range 0x40 {
		scrambled[i] ^= scrambled[0x40+i]
	}

	// Step 3: XOR bytes 0x0D-0x10 with bytes 0x34-0x37
	for i := range 4 {
		scrambled[0x0D+i] ^= scrambled[0x34+i]
	}

	// Step 4: XOR last 0x40 bytes with first 0x40 bytes
	for i := range 0x40 {
		scrambled[0x40+i] ^= scrambled[i]
	}

	return scrambled
}

// RSADecryptNoPadding decrypts a 128-byte block using RSA with no padding
// (RSA/ECB/NoPadding equivalent).
func RSADecryptNoPadding(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) != 128 {
		return nil, fmt.Errorf("RSA decrypt: expected 128 bytes, got %d", len(ciphertext))
	}

	// RSA decrypt without padding: raw RSA operation = ciphertext^d mod n
	c := new(big.Int).SetBytes(ciphertext)
	m := new(big.Int).Exp(c, privateKey.D, privateKey.N)

	result := m.Bytes()
	// Pad to 128 bytes if needed
	if len(result) < 128 {
		padded := make([]byte, 128)
		copy(padded[128-len(result):], result)
		result = padded
	}

	return result, nil
}
