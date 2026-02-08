# L2 Interlude Protocol — Cryptography Deep Dive

> Part 1 of [L2 Interlude Protocol Reference](README.md)
> Previous: [00-overview.md](00-overview.md) | Next: [02-protocol-flows.md](02-protocol-flows.md)

---

## 3. Cryptography Deep Dive

### 3.1 Encryption Layers Overview

The protocol uses three different encryption algorithms at different phases. They are **never nested** — each applies to a specific connection phase.

```mermaid
flowchart TB
    subgraph Phase1["Phase 1: Login Authentication"]
        direction TB
        RSA["RSA-1024\nOne-time password encryption\nClient → LoginServer"]
        BF_STATIC["Blowfish ECB\nStatic key: 6B60CB5B...\nFirst packet only (Init)"]
        BF_DYNAMIC["Blowfish ECB\nDynamic key (per session)\nAll subsequent Login packets"]
        RSA --> BF_STATIC
        BF_STATIC --> BF_DYNAMIC
    end

    subgraph Phase2["Phase 2: Gameplay"]
        XOR["XOR Stream Cipher\n16-byte rolling key\nAll GameServer packets"]
    end

    subgraph Phase3["Internal: GS↔LS Relay"]
        BF_RELAY["Blowfish ECB\nRSA-exchanged key\nAll Relay packets"]
    end

    Phase1 -->|"Client moves to GameServer"| Phase2
    Phase3 ---|"Persistent internal channel"| Phase3
```

### 3.2 RSA-1024 Key Exchange

RSA is used **exactly once** during login — to encrypt the player's password.

#### Key Generation (Server Startup)

- LoginServer generates **10 RSA-1024 key pairs** at startup
- Public exponent: **F4** (65537 = 0x10001)
- Modulus: **128 bytes** (1024 bits)
- A random key pair is selected for each client connection

#### Modulus Scrambling (4-Step Obfuscation)

Before sending the RSA public key to the client, the modulus undergoes a 4-step XOR/swap scrambling process. The client performs the reverse operations to recover the original modulus.

```mermaid
flowchart LR
    A["Original RSA Modulus\n128 bytes (0x00–0x7F)"] --> S1

    S1["Step 1: SWAP\nbytes [0x00–0x03]\n↔ bytes [0x4D–0x50]"] --> S2

    S2["Step 2: XOR\nfor i in 0..0x3F:\n  mod[i] ^= mod[0x40+i]"] --> S3

    S3["Step 3: XOR\nfor i in 0..3:\n  mod[0x0D+i] ^= mod[0x34+i]"] --> S4

    S4["Step 4: XOR\nfor i in 0..0x3F:\n  mod[0x40+i] ^= mod[i]"] --> B

    B["Scrambled Modulus\n→ sent to client in Init packet"]
```

#### RSA Flow

```mermaid
sequenceDiagram
    participant LS as LoginServer
    participant C as Client

    Note over LS: Startup: Generate 10 RSA-1024 key pairs

    LS->>C: Init packet contains:<br/>scrambled RSA modulus (128 bytes)

    Note over C: De-scramble modulus<br/>(reverse 4 steps)

    Note over C: Encrypt login + password<br/>with RSA public key<br/>(RSA/ECB/nopadding)

    C->>LS: RequestAuthLogin<br/>RSA-encrypted block (128 bytes)

    Note over LS: Decrypt with RSA private key<br/>Extract login (offset 0x5E, 14 bytes)<br/>Extract password (offset 0x6C, 16 bytes)
```

### 3.3 Blowfish ECB

Blowfish is the primary encryption for LoginServer packets and the GS↔LS relay channel.

#### Algorithm Parameters

| Parameter | Value |
|-----------|-------|
| Type | Block cipher (Feistel network) |
| Block size | 8 bytes (64 bits) |
| Rounds | 16 |
| P-array | 18 elements (32-bit words) |
| S-boxes | 4 tables × 256 elements each |
| Key schedule | 521 encryption operations |

#### Blowfish Key Schedule

```mermaid
flowchart TD
    A["Start: P-array and S-boxes\nfilled with digits of pi"] --> B["XOR key bytes with P-array\n(cyclic key application)"]
    B --> C["Encrypt (0, 0) → result"]
    C --> D["Write result to P[0], P[1]"]
    D --> E["Encrypt (P[0], P[1]) → result"]
    E --> F["Write result to P[2], P[3]"]
    F --> G["... continue for all 18 P elements\nand 4 x 256 S-box elements"]
    G --> H["Total: 521 encrypt operations\nKey schedule complete"]
```

#### Feistel Round (x16)

```
Each round:
  xL ^= P[i]
  xR ^= F(xL)
  swap(xL, xR)

After 16 rounds:
  xL ^= P[16]
  xR ^= P[17]

F(x) = ((S0[byte3] + S1[byte2]) XOR S2[byte1]) + S3[byte0]
```

#### Checksum (Before Encryption)

Before Blowfish encryption, a 4-byte XOR checksum is appended:

```
checksum = 0
for each 4-byte word in data (except last 4 bytes):
    checksum ^= word
write checksum to last 4 bytes
```

Verification: XOR of all words including checksum = **0**.

#### encXORPass (First Packet Only)

An additional obfuscation layer applied only to the **Init packet**:

```
ecx = random_key  (4-byte int)
for each 4-byte word in data:
    edx = read_int32(data, i)
    ecx += edx
    edx ^= ecx
    write_int32(data, i, edx)
```

This creates a **chain dependency** — each block depends on all previous blocks.

#### Keys Used with Blowfish

| Key | Size | Source | Used For |
|-----|------|--------|----------|
| Static key | 16 bytes | Hardcoded: `6B 60 CB 5B 82 CE 90 B1 CC 2B 6C 55 6C 6C 6C 6C` | First packet (Init) from LoginServer |
| Dynamic key (Login) | 16 bytes | Random from pool of 20 | All Login packets after Init |
| Relay key | 16+ bytes | RSA-encrypted exchange | GS↔LS relay channel |

### 3.4 XOR Stream Cipher

After connecting to GameServer, all packets are encrypted with a lightweight XOR stream cipher.

#### XOR Encryption Formula

```
encrypt: out[i] = raw[i] XOR key[i & 0x0F] XOR out[i-1]
decrypt: raw[i] = enc[i] XOR key[i & 0x0F] XOR enc[i-1]

where out[-1] = 0 (initial value)
```

#### Key Shift (After Each Packet)

```
counter = key[8..11] as uint32_le
counter += packet_size
key[8..11] = counter as bytes_le
```

This ensures identical packets encrypt differently depending on position in the stream.

#### XOR Key Structure

| Bytes | Initial Value | Runtime Role | Description |
|-------|---------------|--------------|-------------|
| 0–7 | Random | XOR key bytes | Generated per connection (1–255 each byte) |
| 8–11 | `C8 27 93 01` | **Counter** | Static at generation; at runtime incremented by packet size after each encrypt/decrypt |
| 12–15 | `A1 6C 31 97` | Static suffix | Never modified during session |

> **Note:** At key pool generation (20 keys), bytes 8–15 are all set to the static suffix `C8 27 93 01 A1 6C 31 97`. At runtime, bytes 8–11 are repurposed as a 32-bit LE counter (key shift), while bytes 12–15 remain constant.

#### XOR State Machine

```mermaid
stateDiagram-v2
    [*] --> Disabled : Connection established

    Disabled --> Passthrough : First encrypt/decrypt call
    note right of Passthrough
        First packet (KeyPacket/ProtocolVersion)
        passes through WITHOUT encryption
    end note

    Passthrough --> Active : Second call onwards
    note right of Active
        All subsequent packets
        encrypted with rolling XOR key
    end note

    Active --> Active : XOR with key, chain with prev byte, shift counter
```

### 3.5 SessionKey Lifecycle

The SessionKey is a **one-time token** (4 x int32 = 128 bits) that links Login and GameServer authentication.

```mermaid
flowchart LR
    A["1. Generate\n(LoginServer)\n4 random int32"] --> B["2. LoginOk\nSend loginOkID1,\nloginOkID2 to client"]
    B --> C["3. PlayOk\nSend playOkID1,\nplayOkID2 to client"]
    C --> D["4. AuthLogin\nClient sends all 4 ints\nto GameServer :7777"]
    D --> E["5. PlayerAuthRequest\nGameServer forwards\nvia Relay :9013"]
    E --> F["6. Validate\nLoginServer compares\nall 4 ints"]
    F --> G["7. PlayerAuthResponse\naccept / reject"]
    G --> H["8. Remove\nSessionKey deleted\nfrom LoginServer"]
```

#### SessionKey Fields

| Field | Size | Set By | Sent In |
|-------|------|--------|---------|
| loginOkID1 | int32 | LoginServer | LoginOk (0x03) |
| loginOkID2 | int32 | LoginServer | LoginOk (0x03) |
| playOkID1 | int32 | LoginServer | PlayOk (0x07) |
| playOkID2 | int32 | LoginServer | PlayOk (0x07) |

### 3.6 LoginEncryption (Two-Stage Blowfish)

LoginServer uses two different Blowfish modes depending on packet order:

```mermaid
flowchart TD
    START([New Login Connection]) --> STATIC["Static Blowfish Key\n6B 60 CB 5B 82 CE 90 B1\nCC 2B 6C 55 6C 6C 6C 6C"]

    STATIC --> FIRST["First Packet (Init)"]
    FIRST --> XOR_PASS["encXORPass(random_int)\nAccumulative XOR obfuscation"]
    XOR_PASS --> BF1["Blowfish ECB Encrypt\n(static key)"]
    BF1 --> SWITCH["Switch to Dynamic Key\n(sent inside Init packet)"]

    SWITCH --> NEXT["All Subsequent Packets"]
    NEXT --> CHECKSUM["appendChecksum()\nXOR of all 4-byte words"]
    CHECKSUM --> BF2["Blowfish ECB Encrypt\n(dynamic key)"]
    BF2 --> NEXT
```

### 3.7 All Keys Summary

| Key | Size | Source | Lifetime | Purpose |
|-----|------|--------|----------|---------|
| RSA private key | 1024 bit | Generated at LS startup (10 pairs) | Process lifetime | Decrypt client password |
| RSA public key (scrambled) | 1024 bit | Sent in Init packet | One session | Client encrypts password |
| Blowfish static key | 16 bytes | Hardcoded in client & server | Permanent | Init packet encryption |
| Blowfish dynamic key (Login) | 16 bytes | Random from pool of 20 | One login session | Login packets after Init |
| Blowfish relay key | 16+ bytes | RSA-exchanged at GS registration | GS↔LS session | All relay packets |
| XOR stream key | 16 bytes | Sent in KeyPacket | One game session | All GameServer packets |
| SessionKey | 4 x int32 | Random, generated by LS | One auth cycle | Link Login→Game auth |
| encXORPass key | 4 bytes (int) | Random | One packet | Init packet obfuscation |

---

> Previous: [00-overview.md](00-overview.md) | Next: [02-protocol-flows.md](02-protocol-flows.md)
