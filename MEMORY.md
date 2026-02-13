# MEMORY.md — Implementation Status

Last updated: 2026-02-13, branch: `try-goroutines`

## Completed Phases

### Phase 1-3: Login & Relay (committed to main)
- LoginServer (TCP :2106) — client authentication, server list
- GS-LS relay (TCP :9013) — GameServer registration, player auth relay

### Phase 4.x: GameServer Core (committed)
- TCP :7777 game client connections
- Packet encryption (Blowfish + GameCrypt)
- Character selection, enter world, leave world
- Logout/restart with 15-second combat delay
- LOD visibility cache, broadcast system
- Hot path optimizations (100K concurrent players)
- Parallel packet encryption for EnterWorld

### Phase 5.x: Gameplay Systems (committed)
- 5.1: Movement validation
- 5.2: Target system with validation
- 5.3: Basic combat (physical auto-attack)
- 5.4: Character templates & stats
- 5.5: Weapon & equipment system
- 5.6: PvE combat (player vs NPC)
- 5.7+5.8: NPC aggro & AI, death/respawn, experience & leveling
- 5.9-6.0: Skills, chat, persistence, data generation

### Phase 7.x: Architecture (committed)
- 7.0: Async write architecture + hot path benchmarks
- 7.1: Shop system, async write, session cleanup, synctest migration

## Current Session — Phase 8: Extended Packet Handlers (uncommitted)

### Implemented Systems

#### Clan System (`handler_clan.go`, `handler_clan_war.go`)
- RequestJoinPledge (0x24) — invite to clan
- RequestAnswerJoinPledge (0x25) — accept/deny
- RequestWithdrawalPledge (0x26) — leave clan
- RequestOustPledgeMember (0x27) — kick from clan
- RequestPledgeInfo (0x66) — clan info
- RequestPledgeCrest (0x68) — clan crest image
- RequestPledgeMemberList (0x3C) — member list
- RequestPledgePower (0xC0) — set rank privileges
- RequestPledgeSetMemberPowerGrade (0xD0:0x1C) — set member rank
- RequestPledgeMemberInfo (0xD0:0x1D) — detailed member info (PledgeReceiveMemberInfo 0xFE:0x3D)
- RequestPledgeReorganizeMember (0xD0:0x24) — move member to sub-pledge
- RequestPledgeWarList (0xD0:0x1E) — war list
- Clan War (0x4D-0x52) — declare/stop/surrender wars

#### Alliance System (`handler_phase50.go`, `handler_phase51.go`)
- RequestAllyInfo (0x8E) — alliance info (AllianceInfo S2C)
- RequestAllyCrest (0x88) — alliance crest image
- RequestSetAllyCrest — upload alliance crest
- RequestJoinAlly / RequestAnswerJoinAlly / RequestDismissAlly / AllyLeave

#### Party System (`handler_party.go`, stubs)
- RequestJoinParty / RequestAnswerJoinParty
- RequestOustPartyMember / RequestWithdrawalParty
- RequestChangePartyLeader (0xD0:0x04)
- Party Matching Room stubs (0xD0:0x00-0x03, 0x14-0x16)

#### Skill System (`handler_phase37.go`, `handler_phase49.go`)
- Skill Learning (0x6B/0x6C) — acquire skill with SP/item cost
- Skill Enchantment (0xD0:0x01-0x02) — enchant skill info and execution
- Skill Cooldown Tracking — real cooldown data via CastManager.GetAllCooldowns()
- RequestSkillCoolTime populated with active cooldowns

#### Combat & Duel (`handler_duel.go`)
- Duel Start / Answer / Surrender (0xD0:0x27-0x29)

#### Social & Chat (`handler_phase36.go`, `handler_friend.go`)
- Social Actions (0x1B → 0x2D)
- Friend system: invite/accept/delete/list/PM (0xCC → 0xFD friend messages)
- Friend PM with block/refusal checks (IsBlocked, MessageRefusal → SystemMessage 176/145)
- Block system: block/unblock/list/allblock/allunblock (0xA0)
- Friend/Block DB persistence via FriendRepository (immediate insert/delete)
- Chat with type validation

#### Item Systems
- Enchant Item (0x58 → 0x81) via `internal/game/enchant` package (scroll data, rates, TryEnchant)
- Use Item dispatch (`handler_use_item.go`, `handler_item_dispatch.go`)
- Recipe system: book open, craft, recipe shop (0xB1-0xB7)
- RequestRecipeBookDestroy (0xAD) — forget recipe
- Augmentation packets (stubs)

#### Respawn System
- Respawn location overrides: castle, clan hall, siege HQ
- Die packet with respawn button flags

#### NPC & World
- Teleporter NPC system
- Henna system (`handler_henna.go`)
- Quest handler (`quest_handler.go`)
- BBS / Community Board (`handler_bbs.go`)
- Link HTML (0x20) + Quest List (0x63)

#### Admin System (`admin/`)
- GM bypass command (0x5B) with IsGM() access level check
- Admin command handler with access level hierarchy
- GM list (0x81)

#### Manor System (fully implemented)
- Manor Manager with period rotation (APPROVED → MAINTENANCE → MODIFIABLE)
- RequestSetSeed (0xD0:0x0A) — set seed production with validation
- RequestSetCrop (0xD0:0x0B) — set crop procurement with validation
- ExShowSeedInfo (0xFE:0x1C) — display seed info
- ExShowCropInfo (0xFE:0x1D) — display crop info
- ExShowSeedSetting (0xFE:0x1F) — seed settings (current + next period)
- ExShowCropSetting (0xFE:0x20) — crop settings (current + next period)
- RequestManorList — castle list for manor UI
- Manor ownership validation via isManorOwner helper

#### Other Stubs
- Olympiad observer/match list (0xD0:0x12-0x13)
- MPCC show party members info (0xD0:0x26)
- Pledge crest large (0xD0:0x10-0x11)
- Seven Signs
- Siege
- Cursed Weapons
- Instance zones
- Clan Hall decorations
- Subclass management
- Pet handler
- Private Store

### New Server Packets (S2C)
- PledgeReceiveMemberInfo (0xFE:0x3D)
- AllianceInfo
- AllyCrest
- SkillCoolTime (populated with real data)
- RecipeBookItemList (on recipe destroy)
- Plus many others from parallel agents

### Key Architecture Decisions
- `CastManager.GetAllCooldowns(objectID)` exposes active cooldowns for SkillCoolTime packet
- GM access check (`player.IsGM()`) at handler level before delegating to admin handler
- Clan/alliance operations use `clan.Table` for thread-safe clan management
- All error wrapping follows `fmt.Errorf("context: %w", err)` pattern
- `slog` for structured logging throughout

## File Count Summary
- 110 modified files + 416 new files vs HEAD (527 total changed)
- Committed to main: 110 files, 6824 insertions, 3382 deletions
- Uncommitted session: 111 files changed, 38454 insertions, 22972 deletions
- New handler files: 25+ handler_*.go files
- New client packets: 80+ parsers
- New server packets: 80+ serializers
- New game systems: 15+ packages under `internal/game/`
