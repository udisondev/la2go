# la2go

Lineage 2 Interlude server emulator written in Go, ported from L2J Mobius CT 0 Interlude Java codebase.

## Quick Start

### Prerequisites

- Go 1.25.7+
- PostgreSQL 17+
- [Task](https://taskfile.dev) (optional, but recommended)

### Installation

```bash
# Clone repository
git clone https://github.com/udisondev/la2go.git
cd la2go/la2go

# Install dependencies
go mod download

# Setup database (PostgreSQL)
# Migrations run automatically on first start

# Copy and configure
cp config/loginserver.yaml.example config/loginserver.yaml
# Edit config/loginserver.yaml with your DB settings

# Build and run
go build -o loginserver ./cmd/loginserver
./loginserver
```

### Configuration

See `config/loginserver.yaml` for configuration options:

- Database connection (PostgreSQL)
- Network settings (bind address, ports)
- LoginServer settings (auto account creation, flood protection, etc.)
- GameServer list

## Development

### Running Tests

```bash
# Show all available tasks
task --list

# Quick unit tests (no DB)
task test-unit

# Integration tests (requires DB)
DB_ADDR='postgres://user:pass@localhost:5432/la2go' task test-integration

# All tests with Docker test DB (automatic setup/teardown)
task test-with-db

# Generate coverage report
task test-coverage
```

### Benchmarking & Profiling

See detailed guide: [docs/BENCHMARKING.md](docs/BENCHMARKING.md)

Quick examples:

```bash
# Quick benchmark run
task bench-quick

# Full benchmark with profiling
task bench-full

# CPU profiling
task profile-cpu

# Compare with baseline
task bench-compare
```

## Project Structure

```
la2go/
├── cmd/
│   └── loginserver/          # LoginServer entry point
├── config/
│   └── loginserver.yaml      # Configuration
├── internal/
│   ├── config/               # Config loading
│   ├── crypto/               # RSA, Blowfish, scrambling
│   ├── db/                   # PostgreSQL, migrations
│   ├── gameserver/           # GameServerTable
│   ├── gslistener/           # GS↔LS relay (TCP :9013)
│   ├── login/                # LoginServer (TCP :2106)
│   ├── model/                # Domain models
│   └── protocol/             # Packet utilities
├── docs/
│   ├── atlas/                # Architecture documentation (27 files)
│   └── BENCHMARKING.md       # Benchmarking guide
└── scripts/
    ├── bench.sh              # Benchmark automation
    └── profile.sh            # Profiling automation
```

## Implementation Status

- ✅ **Phase 1+2: LoginServer** (TCP :2106)
  - Client authentication (RSA + Blowfish)
  - GameGuard handshake
  - Server list
  - Session key generation

- ✅ **Phase 3: GS↔LS Relay** (TCP :9013)
  - GameServer registration
  - Player auth relay
  - SessionKey validation

- ⏳ **Phase 4+: GameServer** (TCP :7777)
  - Not yet implemented

## Documentation

- **Architecture:** [`docs/atlas/`](docs/atlas/) — 27 detailed markdown files covering:
  - Network protocol (crypto, packets, connection flow)
  - LoginServer & GameServer architecture
  - Domain models (player, items, skills, NPCs)
  - Systems (combat, AI, spawn, chat, trade, clans, siege, olympiad, quests)
  - SQL schema, XML datapacks

- **Benchmarking:** [`docs/BENCHMARKING.md`](docs/BENCHMARKING.md) — Complete guide to benchmarking and profiling

- **CLAUDE.md:** Project instructions for AI assistants ([`CLAUDE.md`](../CLAUDE.md))

## Technology Stack

- **Language:** Go 1.25.7
- **Database:** PostgreSQL 17 (pgx driver v5.8.0)
- **Migrations:** goose v3.26.0 (embedded)
- **Config:** YAML (gopkg.in/yaml.v3)
- **Logging:** stdlib `slog`
- **Crypto:** RSA-1024, Blowfish (golang.org/x/crypto/blowfish)
- **Concurrency:** golang.org/x/sync/errgroup

## Protocol

- **L2 Protocol:** Chronicle: Interlude (C6)
- **Protocol revision:** `0x0106`
- **Ports:**
  - 2106 — LoginServer (client)
  - 9013 — GS↔LS relay (internal)
  - 7777 — GameServer (client, not implemented)

## Contributing

See [`CLAUDE.md`](../CLAUDE.md) for development guidelines and architecture overview.

## License

MIT (or L2J Mobius license if applicable)

## Reference

Java codebase: [L2J Mobius CT 0 Interlude](https://github.com/L2jMobius/L2jMobius_CT_0_Interlude)
