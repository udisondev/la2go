package main

import (
	"fmt"
)

func main() {
	// Анализ структуры Init пакета из Java кода:
	// Init.writeImpl():
	//   buffer.writeByte(0x00);              // 1 byte  → offset 0
	//   buffer.writeInt(_sessionId);         // 4 bytes → offset 1
	//   buffer.writeInt(0x0000c621);         // 4 bytes → offset 5
	//   buffer.writeBytes(_publicKey);       // 128 bytes → offset 9
	//   buffer.writeInt(0x29DD954E);         // 4 bytes → offset 137
	//   buffer.writeInt(0x77C39CFC);         // 4 bytes → offset 141
	//   buffer.writeInt(0x97ADB620);         // 4 bytes → offset 145
	//   buffer.writeInt(0x07BDE0F7);         // 4 bytes → offset 149
	//   buffer.writeBytes(_blowfishKey);     // 16 bytes → offset 153
	//   buffer.writeByte(0);                 // 1 byte → offset 169

	fmt.Println("=== Init packet structure (plaintext) ===")
	fmt.Println("Offset   Size   Field")
	fmt.Println("------   ----   -----")

	offset := 0
	fmt.Printf("%6d   %4d   opcode (0x00)\n", offset, 1)
	offset += 1

	fmt.Printf("%6d   %4d   sessionID\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   protocolRevision (0x0000c621)\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   rsaPublicKey (scrambled modulus)\n", offset, 128)
	offset += 128

	fmt.Printf("%6d   %4d   ggData[0] (0x29DD954E)\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   ggData[1] (0x77C39CFC)\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   ggData[2] (0x97ADB620)\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   ggData[3] (0x07BDE0F7)\n", offset, 4)
	offset += 4

	fmt.Printf("%6d   %4d   blowfishKey\n", offset, 16)
	blowfishKeyOffset := offset
	offset += 16

	fmt.Printf("%6d   %4d   null termination\n", offset, 1)
	offset += 1

	plaintextSize := offset

	fmt.Printf("\nTotal plaintext size: %d bytes\n", plaintextSize)
	fmt.Printf("Blowfish key offset: %d (0x%X)\n", blowfishKeyOffset, blowfishKeyOffset)

	// Расчёт encryptedSize по формуле Java:
	// LoginEncryption.encryptedSize(dataSize):
	//   dataSize += _static ? 8 : 4;
	//   dataSize += 8 - (dataSize % 8);
	//   dataSize += 8;
	//   return dataSize;

	dataSize := plaintextSize
	fmt.Printf("\n=== encryptedSize calculation (Java formula) ===\n")
	fmt.Printf("Initial dataSize: %d\n", dataSize)

	dataSize += 8 // _static=true
	fmt.Printf("After +8 (static): %d\n", dataSize)

	padding := 8 - (dataSize % 8)
	if padding == 8 {
		padding = 0
	}
	dataSize += padding
	fmt.Printf("After padding to 8: %d (padding=%d)\n", dataSize, padding)

	dataSize += 8
	fmt.Printf("After +8 (final): %d\n", dataSize)

	encryptedSize := dataSize

	fmt.Printf("\nFinal encrypted size: %d bytes\n", encryptedSize)

	// encXORPass диапазон:
	// encXORPass(data, offset=0, size=encryptedSize, key=randomInt):
	//   stop = size - 8
	//   pos = 4 + offset (starts at 4, skips first 4 bytes)
	//   while pos < stop: XOR process
	//   writeInt(pos, ecx) — writes final accumulated key

	fmt.Printf("\n=== encXORPass processing ===\n")
	fmt.Printf("encXORPass(data, offset=0, size=%d, key=randomInt)\n", encryptedSize)
	fmt.Printf("  start (pos=4+offset): 4\n")
	fmt.Printf("  stop (size-8): %d\n", encryptedSize-8)
	fmt.Printf("  final XOR key written at: %d--%d\n", encryptedSize-8, encryptedSize-4)

	fmt.Printf("\n=== Blowfish key position ===\n")
	fmt.Printf("Blowfish key range: %d--%d (16 bytes)\n", blowfishKeyOffset, blowfishKeyOffset+16)
	fmt.Printf("encXORPass processes: 4--%d\n", encryptedSize-8)

	if blowfishKeyOffset >= 4 && blowfishKeyOffset+16 <= encryptedSize-8 {
		fmt.Printf("\n✅ Blowfish key IS encrypted by encXORPass (range: 4--%d includes %d--%d)\n",
			encryptedSize-8, blowfishKeyOffset, blowfishKeyOffset+16)
	} else {
		fmt.Printf("\n❌ Blowfish key is NOT encrypted by encXORPass\n")
	}

	// decXORPass должен обработать те же диапазоны
	fmt.Printf("\n=== decXORPass on client ===\n")
	fmt.Printf("decXORPass(data, offset=0, size=%d)\n", encryptedSize)
	fmt.Printf("  stop = offset + size - 8 = %d\n", encryptedSize-8)
	fmt.Printf("  pos = offset + 4 = 4\n")
	fmt.Printf("  ecx = data[stop] (read final XOR key from %d)\n", encryptedSize-8)
	fmt.Printf("  for i = stop-4 down to pos (i.e., %d down to 4, step -4):\n", encryptedSize-8-4)
	fmt.Printf("    edx = data[i]\n")
	fmt.Printf("    edx ^= ecx\n")
	fmt.Printf("    data[i] = edx\n")
	fmt.Printf("    ecx -= edx\n")

	fmt.Printf("\n=== CRITICAL ===\n")
	fmt.Printf("After Blowfish decrypt AND decXORPass:\n")
	fmt.Printf("  Blowfish key will be at offset %d (plaintext)\n", blowfishKeyOffset)
	fmt.Printf("  Client must extract key from offset %d--%d (16 bytes)\n", blowfishKeyOffset, blowfishKeyOffset+16)
}
