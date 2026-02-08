# L2 Interlude Protocol — Complete Reference

> Comprehensive visual guide to the Lineage 2 Chronicle: Interlude (C6) network protocol.
> Based on L2J Mobius CT 0 Interlude Java codebase and la2go Go implementation.

---

## Documents

| # | File | Sections | Description |
|---|------|----------|-------------|
| 0 | [00-overview.md](00-overview.md) | 1–2 | Executive Summary, Architecture Overview |
| 1 | [01-crypto.md](01-crypto.md) | 3 | Cryptography Deep Dive (RSA, Blowfish, XOR, SessionKey) |
| 2 | [02-protocol-flows.md](02-protocol-flows.md) | 4–6 | Sequence Diagrams, State Machines (FSM), Use Case Scenarios |
| 3 | [03-packet-reference.md](03-packet-reference.md) | 7–8 | Packet Reference Tables, Binary Packet Structures |
| 4 | [04-design-and-appendix.md](04-design-and-appendix.md) | 9–12 | Protocol Design Principles, Comparative Tables, Security, Appendix |

## Protocol at a Glance

| Metric | Value |
|--------|-------|
| Total packets | **476** (177 C→S + 269 S→C + 17 Login + 13 GS↔LS Relay) |
| Encryption layers | **3** (RSA-1024, Blowfish ECB, XOR Stream) |
| Server processes | **3** (LoginServer, GameServer, GS↔LS Relay) |
| TCP ports | **3** (:2106 Login, :7777 Game, :9013 Relay) |
| Protocol revision | `0x0000C621` (Interlude) |
| Byte order | **Little-Endian** throughout |
| String encoding | **UTF-16LE** null-terminated |

---

*Generated for the la2go project — L2 Interlude server emulator in Go.*
*Based on L2J Mobius CT 0 Interlude Java codebase (1619 files) and la2go Go implementation.*
