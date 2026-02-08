package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"

	"github.com/udisondev/la2go/internal/constants"
)

// RSAKeyPair holds an RSA-1024 key pair and the scrambled modulus for the client.
type RSAKeyPair struct {
	PrivateKey       *rsa.PrivateKey
	ScrambledModulus []byte // 128 bytes, scrambled for L2 client
}

// GenerateRSAKeyPair generates an RSA-1024 key pair with exponent 65537 (F4)
// and pre-computes the scrambled modulus for Client↔LoginServer protocol.
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, constants.RSAKeyBits)
	if err != nil {
		return nil, fmt.Errorf("generating RSA key: %w", err)
	}

	// Pre-compute CRT values (Dp, Dq, Qinv) to enable Chinese Remainder Theorem optimizations
	// in crypto/rsa.DecryptPKCS1v15 and raw RSA operations. This provides ~20-30% speedup.
	privateKey.Precompute()

	modBytes := privateKey.PublicKey.N.Bytes()

	// Java BigInteger.toByteArray() may return 129 bytes with leading zero.
	// We need exactly 128 bytes.
	if len(modBytes) == constants.RSAModulusMaxSize && modBytes[0] == 0 {
		modBytes = modBytes[1:]
	}
	if len(modBytes) < constants.RSA1024ModulusSize {
		padded := make([]byte, constants.RSA1024ModulusSize)
		copy(padded[constants.RSA1024ModulusSize-len(modBytes):], modBytes)
		modBytes = padded
	}

	scrambled := ScrambleModulus(modBytes)

	return &RSAKeyPair{
		PrivateKey:       privateKey,
		ScrambledModulus: scrambled,
	}, nil
}

// GenerateRSAKeyPair512 generates an RSA-512 key pair with exponent 65537 (F4)
// for GameServer↔LoginServer protocol. Returns raw modulus (no scrambling).
//
// Note: Go 1.19+ blocks generation of RSA keys < 1024 bits as insecure.
// For L2 protocol compatibility (Java original uses RSA-512), we need a workaround.
// In production, pre-generate keys externally and load them.
func GenerateRSAKeyPair512() (*RSAKeyPair, error) {
	// Workaround: Используем фиксированный тестовый ключ RSA-512
	// В production нужно сгенерировать ключи заранее (например, через OpenSSL)
	// и загружать их из файла.

	// Используем тестовый ключ из testRSA512Key()
	privateKey := testRSA512Key()

	// Pre-compute CRT values for faster RSA operations
	privateKey.Precompute()

	modBytes := privateKey.PublicKey.N.Bytes()

	// RSA-512 → 64 bytes expected
	// Java BigInteger.toByteArray() may return 65 bytes with leading zero.
	if len(modBytes) == constants.RSA512ModulusSize+1 && modBytes[0] == 0 {
		modBytes = modBytes[1:]
	}
	if len(modBytes) < constants.RSA512ModulusSize {
		padded := make([]byte, constants.RSA512ModulusSize)
		copy(padded[constants.RSA512ModulusSize-len(modBytes):], modBytes)
		modBytes = padded
	}

	// GS↔LS protocol: no scrambling, just raw modulus
	return &RSAKeyPair{
		PrivateKey:       privateKey,
		ScrambledModulus: modBytes, // raw modulus, not scrambled
	}, nil
}

// testRSA512Key returns a fixed RSA-512 test key.
// Generated externally using: openssl genrsa 512
func testRSA512Key() *rsa.PrivateKey {
	// Modulus N (512-bit, exactly 64 bytes with MSB set)
	n := new(big.Int)
	n.SetString("a7f58ef05452ac91062310847dba84f92168437ff032fea96c2df71c2f62b80ca6130ab1aeb861d0e28acba3dec82965803e81ad1dd09d331816c8bd9647e31b", 16)

	// Public exponent E
	e := constants.RSAPublicExponent

	// Private exponent D
	d := new(big.Int)
	d.SetString("9e391ba0972f12d5c3bc40912f88084051125194328937920f10f61b1d209853fb3df6f6dc22cddf46e182585f95a9985b4f4a5530f1a8591ee0cffb80ae8a61", 16)

	// Primes P and Q
	p := new(big.Int)
	p.SetString("d57bbc1f292819d705f4228509f0efa56d577f0806969a89fc47ae7df41235b3", 16)

	q := new(big.Int)
	q.SetString("c968d09cf98b4c05d1350550bf781fa4bc6a3df2690f3aab449ea24ff3a3b8f9", 16)

	return &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: n,
			E: e,
		},
		D:      d,
		Primes: []*big.Int{p, q},
	}
}

// ScrambleModulus applies the 4-step XOR/swap obfuscation to the RSA modulus
// as done in L2J ScrambledKeyPair.java.
// Input must be exactly 128 bytes.
func ScrambleModulus(modulus []byte) []byte {
	if len(modulus) != constants.RSA1024ModulusSize {
		panic(fmt.Sprintf("ScrambleModulus: expected %d bytes, got %d", constants.RSA1024ModulusSize, len(modulus)))
	}

	scrambled := make([]byte, constants.RSA1024ModulusSize)
	copy(scrambled, modulus)

	// Step 1: swap bytes 0x00-0x03 with 0x4D-0x50
	for i := range constants.ScrambleSwapLength {
		scrambled[constants.ScrambleSwapOffset1+i], scrambled[constants.ScrambleSwapOffset2+i] =
			scrambled[constants.ScrambleSwapOffset2+i], scrambled[constants.ScrambleSwapOffset1+i]
	}

	// Step 2: XOR first 0x40 bytes with last 0x40 bytes
	for i := range constants.ScrambleXORBlock1Size {
		scrambled[constants.ScrambleXORBlock1Start+i] ^= scrambled[constants.ScrambleXORBlock2Start+i]
	}

	// Step 3: XOR bytes 0x0D-0x10 with bytes 0x34-0x37
	for i := range constants.ScrambleXORLength {
		scrambled[constants.ScrambleXOROffset1+i] ^= scrambled[constants.ScrambleXOROffset2+i]
	}

	// Step 4: XOR last 0x40 bytes with first 0x40 bytes
	for i := range constants.ScrambleXORBlock1Size {
		scrambled[constants.ScrambleXORBlock2Start+i] ^= scrambled[constants.ScrambleXORBlock1Start+i]
	}

	return scrambled
}

// UnscrambleModulus reverses the ScrambleModulus operation to restore the original modulus.
// Client uses this to extract the original RSA public key from the scrambled modulus in Init packet.
// Input must be exactly 128 bytes.
func UnscrambleModulus(scrambled []byte) []byte {
	if len(scrambled) != constants.RSA1024ModulusSize {
		panic(fmt.Sprintf("UnscrambleModulus: expected %d bytes, got %d", constants.RSA1024ModulusSize, len(scrambled)))
	}

	unscrambled := make([]byte, constants.RSA1024ModulusSize)
	copy(unscrambled, scrambled)

	// Apply operations in REVERSE order

	// Step 4 reverse: XOR last 0x40 bytes with first 0x40 bytes
	for i := range constants.ScrambleXORBlock1Size {
		unscrambled[constants.ScrambleXORBlock2Start+i] ^= unscrambled[constants.ScrambleXORBlock1Start+i]
	}

	// Step 3 reverse: XOR bytes 0x0D-0x10 with bytes 0x34-0x37
	for i := range constants.ScrambleXORLength {
		unscrambled[constants.ScrambleXOROffset1+i] ^= unscrambled[constants.ScrambleXOROffset2+i]
	}

	// Step 2 reverse: XOR first 0x40 bytes with last 0x40 bytes
	for i := range constants.ScrambleXORBlock1Size {
		unscrambled[constants.ScrambleXORBlock1Start+i] ^= unscrambled[constants.ScrambleXORBlock2Start+i]
	}

	// Step 1 reverse: swap bytes 0x00-0x03 with 0x4D-0x50
	for i := range constants.ScrambleSwapLength {
		unscrambled[constants.ScrambleSwapOffset1+i], unscrambled[constants.ScrambleSwapOffset2+i] =
			unscrambled[constants.ScrambleSwapOffset2+i], unscrambled[constants.ScrambleSwapOffset1+i]
	}

	return unscrambled
}

// RSADecryptNoPadding decrypts a block using RSA with no padding (RSA/ECB/NoPadding).
//
// SECURITY NOTES:
// - Uses CRT (Chinese Remainder Theorem) for 2.6x speedup when Precomputed values available
// - NOT constant-time: CRT path ~115µs vs fallback ~298µs (timing leak)
// - Acceptable for L2 login protocol (one-shot operation, legacy protocol)
// - For security-critical applications, consider constant-time wrapper or crypto/rsa.DecryptOAEP
//
// CRT Algorithm (Garner's):
//   m1 = c^dP mod p
//   m2 = c^dQ mod q
//   h = (m1 - m2) * qInv mod p
//   m = m2 + h*q
//
// Expected ciphertext size:
// - RSA-512: 64 bytes
// - RSA-1024: 128 bytes
func RSADecryptNoPadding(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	// Определяем ожидаемый размер по размеру ключа
	keySize := privateKey.N.BitLen() / 8

	if len(ciphertext) != keySize {
		return nil, fmt.Errorf("RSA decrypt: expected %d bytes for %d-bit key, got %d", keySize, privateKey.N.BitLen(), len(ciphertext))
	}

	c := new(big.Int).SetBytes(ciphertext)

	// CRT optimization: if Precomputed values are available, use Chinese Remainder Theorem
	// for 2.6x speedup. Algorithm from Go stdlib crypto/rsa (Garner's algorithm).
	// All three CRT components (Dp, Dq, Qinv) must be present for safe CRT usage.
	if privateKey.Precomputed.Dp != nil &&
		privateKey.Precomputed.Dq != nil &&
		privateKey.Precomputed.Qinv != nil &&
		len(privateKey.Primes) >= 2 {
		// m1 = c^dP mod p
		m1 := new(big.Int).Exp(c, privateKey.Precomputed.Dp, privateKey.Primes[0])

		// m2 = c^dQ mod q
		m2 := new(big.Int).Exp(c, privateKey.Precomputed.Dq, privateKey.Primes[1])

		// h = (m1 - m2) * qInv mod p
		h := new(big.Int).Sub(m1, m2)
		h.Mul(h, privateKey.Precomputed.Qinv)
		h.Mod(h, privateKey.Primes[0])

		// m = m2 + h*q
		m := new(big.Int).Mul(h, privateKey.Primes[1])
		m.Add(m, m2)

		result := m.Bytes()
		if len(result) < keySize {
			padded := make([]byte, keySize)
			copy(padded[keySize-len(result):], result)
			result = padded
		}
		return result, nil
	}

	// Fallback: raw RSA operation = ciphertext^d mod n (slower)
	m := new(big.Int).Exp(c, privateKey.D, privateKey.N)

	result := m.Bytes()
	// Pad to keySize bytes if needed
	if len(result) < keySize {
		padded := make([]byte, keySize)
		copy(padded[keySize-len(result):], result)
		result = padded
	}

	return result, nil
}
