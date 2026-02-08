# L2 Interlude Protocol — Packet Reference

> Part 3 of [L2 Interlude Protocol Reference](README.md)
> Previous: [02-protocol-flows.md](02-protocol-flows.md) | Next: [04-design-and-appendix.md](04-design-and-appendix.md)

---

## 7. Packet Reference Tables

### 7.1 LoginServer Packets

#### Client → Server (6 packets)

| Opcode | Name | Format | Size | State | Description |
|--------|------|--------|------|-------|-------------|
| `0x00` | RequestAuthLogin | RSA block | 128/256 | AUTHED_GG | Login + password (RSA encrypted) |
| `0x02` | RequestServerLogin | `ddc` | 9 | AUTHED_LOGIN | Select game server (skey1, skey2, serverId) |
| `0x05` | RequestServerList | `dd` | 8 | AUTHED_LOGIN | Request server list (skey1, skey2) |
| `0x07` | AuthGameGuard | `ddddd` | 20 | CONNECTED | GameGuard session (sessionId + 4x data) |
| `0x0E` | RequestPIAgreementCheck | — | — | AUTHED_LOGIN | PI agreement check (not implemented) |
| `0x0F` | RequestPIAgreement | — | — | AUTHED_LOGIN | PI agreement (not implemented) |

#### Server → Client (11 packets)

| Opcode | Name | Format | Size | Description |
|--------|------|--------|------|-------------|
| `0x00` | Init | `c d d b[128] b[16] d d d d b[16] c` | 186 | RSA key + BF key + GG constants |
| `0x01` | LoginFail | `cc` | 2 | Error reason code |
| `0x02` | AccountKicked | `cd` | 5 | Account kicked (reason int32) |
| `0x03` | LoginOk | `c d d d d d d d d b[16]` | 49 | loginOkID1, loginOkID2 + padding |
| `0x04` | ServerList | variable | var | Server list with char counts |
| `0x06` | PlayFail | `cc` | 2 | Server selection error |
| `0x07` | PlayOk | `cdd` | 9 | playOkID1, playOkID2 |
| `0x0B` | GGAuth | `cddddd` | 21 | GameGuard response |
| `0x0D` | LoginOptFail | — | — | Optional check fail |
| `0x11` | PIAgreementCheck | — | — | PI agreement check |
| `0x12` | PIAgreementAck | — | — | PI agreement ack |

### 7.2 GameServer Packets — Client → Server

#### Authentication (4 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x00` | ProtocolVersion | CONNECTED | Protocol revision (0xC621) |
| `0x08` | AuthLogin | CONNECTED | Account + SessionKey (4x int32) |
| `0x09` | Logout | AUTHENTICATED, IN_GAME | Request logout |
| `0x46` | RequestRestart | IN_GAME | Return to character selection |

#### Character Management (5 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x0B` | CharacterCreate | AUTHENTICATED | Create new character |
| `0x0C` | CharacterDelete | AUTHENTICATED | Delete character |
| `0x0D` | CharacterSelect | AUTHENTICATED | Select character to play |
| `0x0E` | NewCharacter | AUTHENTICATED | Request class templates |
| `0x62` | CharacterRestore | AUTHENTICATED | Restore deleted character |

#### Movement (6 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x01` | MoveToLocation | IN_GAME | Request movement to (x,y,z) |
| `0x36` | CannotMoveAnymore | IN_GAME | Client reports movement blocked |
| `0x41` | MoveWithDelta | IN_GAME | Relative movement |
| `0x48` | ValidatePosition | IN_GAME | Client position validation |
| `0x4A` | StartRotating | IN_GAME | Begin rotation |
| `0x4B` | FinishRotating | IN_GAME | End rotation |

#### Combat & Actions (8 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x04` | Action | IN_GAME | Click on object/NPC |
| `0x0A` | AttackRequest | IN_GAME | Attack target |
| `0x1B` | RequestSocialAction | IN_GAME | Emote/social action |
| `0x1C` | ChangeMoveType2 | IN_GAME | Walk/run toggle |
| `0x1D` | ChangeWaitType2 | IN_GAME | Sit/stand toggle |
| `0x2F` | RequestMagicSkillUse | IN_GAME | Use skill |
| `0x37` | RequestTargetCanceld | IN_GAME | Cancel target |
| `0x45` | RequestActionUse | IN_GAME | Use action (pet/summon) |

#### Inventory & Items (6 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x0F` | RequestItemList | IN_GAME | Request full inventory |
| `0x11` | RequestUnEquipItem | IN_GAME | Unequip item |
| `0x12` | RequestDropItem | IN_GAME | Drop item on ground |
| `0x14` | UseItem | IN_GAME | Use/equip item |
| `0x59` | RequestDestroyItem | IN_GAME | Destroy item |
| `0x72` | RequestCrystallizeItem | IN_GAME | Crystallize equipment |

#### Trade (6 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x15` | TradeRequest | IN_GAME | Initiate trade |
| `0x16` | AddTradeItem | IN_GAME | Add item to trade window |
| `0x17` | TradeDone | IN_GAME | Confirm/cancel trade |
| `0x1E` | RequestSellItem | IN_GAME | Sell to NPC |
| `0x1F` | RequestBuyItem | IN_GAME | Buy from NPC |
| `0x44` | AnswerTradeRequest | IN_GAME | Accept/reject trade request |

#### Chat (2 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x38` | Say2 | IN_GAME | Send chat message (all types) |
| `0xCC` | RequestSendFriendMsg | IN_GAME | Send private message to friend |

#### Party (4 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x29` | RequestJoinParty | IN_GAME | Invite to party |
| `0x2A` | RequestAnswerJoinParty | IN_GAME | Accept/reject party invite |
| `0x2B` | RequestWithDrawalParty | IN_GAME | Leave party |
| `0x2C` | RequestOustPartyMember | IN_GAME | Kick party member |

#### Clan (11 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x24` | RequestJoinPledge | IN_GAME | Invite to clan |
| `0x25` | RequestAnswerJoinPledge | IN_GAME | Accept/reject clan invite |
| `0x26` | RequestWithdrawalPledge | IN_GAME | Leave clan |
| `0x27` | RequestOustPledgeMember | IN_GAME | Kick clan member |
| `0x3C` | RequestPledgeMemberList | IN_GAME | Request member list |
| `0x53` | RequestSetPledgeCrest | IN_GAME | Set clan crest |
| `0x55` | RequestGiveNickName | IN_GAME | Set clan title |
| `0x66` | RequestPledgeInfo | IN_GAME | Request clan info |
| `0x67` | RequestPledgeExtendedInfo | IN_GAME | Request extended clan info |
| `0x68` | RequestPledgeCrest | AUTH, IN_GAME | Request clan crest image |
| `0xC0` | RequestPledgePower | IN_GAME | Manage clan powers |

#### Skills (4 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x3F` | RequestSkillList | IN_GAME | Request skill list |
| `0x6B` | RequestAcquireSkillInfo | IN_GAME | Skill learning info |
| `0x6C` | RequestAcquireSkill | IN_GAME | Learn skill |
| `0x9D` | RequestSkillCoolTime | IN_GAME | Request cooldowns |

#### Alliance (8 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x82` | RequestJoinAlly | IN_GAME | Invite to alliance |
| `0x83` | RequestAnswerJoinAlly | IN_GAME | Accept/reject |
| `0x84` | AllyLeave | IN_GAME | Leave alliance |
| `0x85` | AllyDismiss | IN_GAME | Dismiss alliance |
| `0x86` | RequestDismissAlly | IN_GAME | Request dismiss |
| `0x87` | RequestSetAllyCrest | IN_GAME | Set alliance crest |
| `0x88` | RequestAllyCrest | IN_GAME | Get alliance crest |
| `0x8E` | RequestAllyInfo | IN_GAME | Request alliance info |

#### Friends (4 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x5E` | RequestFriendInvite | IN_GAME | Add friend |
| `0x5F` | RequestAnswerFriendInvite | IN_GAME | Accept/reject friend request |
| `0x60` | RequestFriendList | IN_GAME | Request friend list |
| `0x61` | RequestFriendDel | IN_GAME | Remove friend |

#### Clan Wars (6 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x4D` | RequestStartPledgeWar | IN_GAME | Declare clan war |
| `0x4E` | RequestReplyStartPledgeWar | IN_GAME | Reply to war declaration |
| `0x4F` | RequestStopPledgeWar | IN_GAME | End clan war |
| `0x50` | RequestReplyStopPledgeWar | IN_GAME | Reply to end war |
| `0x51` | RequestSurrenderPledgeWar | IN_GAME | Surrender |
| `0x52` | RequestReplySurrenderPledgeWar | IN_GAME | Reply to surrender |

#### Private Store (10 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x73` | RequestPrivateStoreManageSell | IN_GAME | Open sell store setup |
| `0x74` | SetPrivateStoreListSell | IN_GAME | Set sell list |
| `0x76` | RequestPrivateStoreQuitSell | IN_GAME | Close sell store |
| `0x77` | SetPrivateStoreMsgSell | IN_GAME | Set sell message |
| `0x79` | RequestPrivateStoreBuy | IN_GAME | Buy from private store |
| `0x90` | RequestPrivateStoreManageBuy | IN_GAME | Open buy store setup |
| `0x91` | SetPrivateStoreListBuy | IN_GAME | Set buy list |
| `0x93` | RequestPrivateStoreQuitBuy | IN_GAME | Close buy store |
| `0x94` | SetPrivateStoreMsgBuy | IN_GAME | Set buy message |
| `0x96` | RequestPrivateStoreSell | IN_GAME | Sell to private store |

#### Pet / Summon (5 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x89` | RequestChangePetName | IN_GAME | Rename pet |
| `0x8A` | RequestPetUseItem | IN_GAME | Pet uses item |
| `0x8B` | RequestGiveItemToPet | IN_GAME | Give item to pet |
| `0x8C` | RequestGetItemFromPet | IN_GAME | Take item from pet |
| `0x8F` | RequestPetGetItem | IN_GAME | Pet picks up item |

#### Vehicle (4 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x42` | RequestGetOnVehicle | IN_GAME | Board vehicle |
| `0x43` | RequestGetOffVehicle | IN_GAME | Leave vehicle |
| `0x5C` | RequestMoveToLocationInVehicle | IN_GAME | Move while on vehicle |
| `0x5D` | CannotMoveAnymoreInVehicle | IN_GAME | Can't move on vehicle |

#### Siege (5 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x47` | RequestSiegeInfo | IN_GAME | Siege info |
| `0xA2` | RequestSiegeAttackerList | IN_GAME | Attacker list |
| `0xA3` | RequestSiegeDefenderList | IN_GAME | Defender list |
| `0xA4` | RequestJoinSiege | IN_GAME | Join siege |
| `0xA5` | RequestConfirmSiegeWaitingList | IN_GAME | Confirm waiting list |

#### Recipe / Craft (10 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0xAC` | RequestRecipeBookOpen | IN_GAME | Open recipe book |
| `0xAD` | RequestRecipeBookDestroy | IN_GAME | Delete recipe |
| `0xAE` | RequestRecipeItemMakeInfo | IN_GAME | Craft info |
| `0xAF` | RequestRecipeItemMakeSelf | IN_GAME | Craft item |
| `0xB1` | RequestRecipeShopMessageSet | IN_GAME | Set shop message |
| `0xB2` | RequestRecipeShopListSet | IN_GAME | Set shop recipe list |
| `0xB3` | RequestRecipeShopManageQuit | IN_GAME | Close recipe shop |
| `0xB5` | RequestRecipeShopMakeInfo | IN_GAME | Shop craft info |
| `0xB6` | RequestRecipeShopMakeItem | IN_GAME | Shop craft item |
| `0xB7` | RequestRecipeShopManagePrev | IN_GAME | Previous page |

#### Henna (6 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0xBA` | RequestHennaItemList | IN_GAME | Available hennas |
| `0xBB` | RequestHennaItemInfo | IN_GAME | Henna details |
| `0xBC` | RequestHennaEquip | IN_GAME | Apply henna |
| `0xBD` | RequestHennaRemoveList | IN_GAME | Hennas to remove |
| `0xBE` | RequestHennaItemRemoveInfo | IN_GAME | Henna remove info |
| `0xBF` | RequestHennaRemove | IN_GAME | Remove henna |

#### Warehouse (2 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x31` | SendWareHouseDepositList | IN_GAME | Deposit items |
| `0x32` | SendWareHouseWithDrawList | IN_GAME | Withdraw items |

#### Quests (2 packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x63` | RequestQuestList | IN_GAME | Request quest list |
| `0x64` | RequestQuestAbort | IN_GAME | Abandon quest |

#### Other (remaining packets)

| Opcode | Name | State | Description |
|--------|------|-------|-------------|
| `0x20` | RequestLinkHtml | IN_GAME | NPC HTML link |
| `0x21` | RequestBypassToServer | IN_GAME | NPC bypass command |
| `0x30` | Appearing | IN_GAME | Character appeared |
| `0x33` | RequestShortCutReg | IN_GAME | Register shortcut |
| `0x35` | RequestShortCutDel | IN_GAME | Delete shortcut |
| `0x57` | RequestShowBoard | IN_GAME | Community board |
| `0x58` | RequestEnchantItem | IN_GAME | Enchant item |
| `0x6D` | RequestRestartPoint | IN_GAME | Respawn point selection |
| `0x6E` | RequestGMCommand | IN_GAME | GM command |
| `0x7B`–`0x7E` | Tutorial* | IN_GAME | Tutorial packets (4) |
| `0x7F`–`0x80` | Petition* | IN_GAME | Petition packets (2) |
| `0x81` | RequestGmList | IN_GAME | Request GM list |
| `0xA0` | RequestBlock | IN_GAME | Block player |
| `0xA7` | MultiSellChoose | IN_GAME | Multi-sell selection |
| `0xA8` | NetPing | IN_GAME | Network ping response |
| `0xAA` | BypassUserCmd | IN_GAME | User command bypass |
| `0xB8` | ObserverReturn | IN_GAME | Return from observation |
| `0xB9` | RequestEvaluate | IN_GAME | Evaluate player |
| `0xC1` | RequestMakeMacro | IN_GAME | Create macro |
| `0xC2` | RequestDeleteMacro | IN_GAME | Delete macro |
| `0xC4` | RequestBuySeed | IN_GAME | Buy seed (manor) |
| `0xC5` | DlgAnswer | IN_GAME | Dialog answer |
| `0xC6` | RequestPreviewItem | IN_GAME | Preview item |
| `0xC7` | RequestSSQStatus | IN_GAME | Seven Signs status |
| `0xCA` | GameGuardReply | IN_GAME | GameGuard reply |
| `0xCD` | RequestShowMiniMap | IN_GAME | Show mini map |
| `0xCF` | RequestRecordInfo | IN_GAME | Record info |
| `0xD0` | ExPacket | all states | Extended packet marker |

### 7.3 GameServer Packets — Client → Server (Extended, prefix 0xD0)

| Sub-opcode | Name | Category |
|------------|------|----------|
| `0x01` | RequestOustFromPartyRoom | Party Room |
| `0x02` | RequestDismissPartyRoom | Party Room |
| `0x03` | RequestWithdrawPartyRoom | Party Room |
| `0x04` | RequestChangePartyLeader | Party |
| `0x05` | RequestAutoSoulShot | Combat |
| `0x06` | RequestExEnchantSkillInfo | Skills |
| `0x07` | RequestExEnchantSkill | Skills |
| `0x08` | RequestManorList | Manor |
| `0x09` | RequestProcureCropList | Manor |
| `0x0A` | RequestSetSeed | Manor |
| `0x0B` | RequestSetCrop | Manor |
| `0x0C` | RequestWriteHeroWords | Hero |
| `0x0D` | RequestExAskJoinMPCC | MPCC |
| `0x0E` | RequestExAcceptJoinMPCC | MPCC |
| `0x0F` | RequestExOustFromMPCC | MPCC |
| `0x10` | RequestExPledgeCrestLarge | Clan |
| `0x11` | RequestExSetPledgeCrestLarge | Clan |
| `0x12` | RequestOlympiadObserverEnd | Olympiad |
| `0x13` | RequestOlympiadMatchList | Olympiad |
| `0x14` | RequestAskJoinPartyRoom | Party Room |
| `0x15` | AnswerJoinPartyRoom | Party Room |
| `0x16` | RequestListPartyMatchingWaitingRoom | Party |
| `0x17` | RequestExitPartyMatchingWaitingRoom | Party |
| `0x18` | RequestGetBossRecord | Boss |
| `0x19` | RequestPledgeSetAcademyMaster | Clan |
| `0x1A` | RequestPledgePowerGradeList | Clan |
| `0x1B` | RequestPledgeMemberPowerInfo | Clan |
| `0x1C` | RequestPledgeSetMemberPowerGrade | Clan |
| `0x1D` | RequestPledgeMemberInfo | Clan |
| `0x1E` | RequestPledgeWarList | Clan War |
| `0x1F` | RequestExFishRanking | Fishing |
| `0x20` | RequestPCCafeCouponUse | PC Cafe |
| `0x22` | RequestCursedWeaponList | Cursed |
| `0x23` | RequestCursedWeaponLocation | Cursed |
| `0x24` | RequestPledgeReorganizeMember | Clan |
| `0x26` | RequestExMPCCShowPartyMembersInfo | MPCC |
| `0x27` | RequestDuelStart | Duel |
| `0x28` | RequestDuelAnswerStart | Duel |
| `0x29` | RequestConfirmTargetItem | Augment |
| `0x2A` | RequestConfirmRefinerItem | Augment |
| `0x2B` | RequestConfirmGemStone | Augment |
| `0x2C` | RequestRefine | Augment |
| `0x2D` | RequestConfirmCancelItem | Augment |
| `0x2E` | RequestRefineCancel | Augment |
| `0x2F` | RequestExMagicSkillUseGround | Skills |
| `0x30` | RequestDuelSurrender | Duel |

### 7.4 GS↔LS Relay Packets

#### GameServer → LoginServer (6 packets)

| Name | Description |
|------|-------------|
| BlowFishKey | RSA-encrypted Blowfish key (64 bytes) |
| GameServerAuth | hexID, port, maxPlayers, subnets, hosts |
| PlayerAuthRequest | account + SessionKey (4x int32) |
| PlayerInGame | List of online account names |
| PlayerLogout | Account name that logged out |
| ServerStatus | Current/max players, status flags |

#### LoginServer → GameServer (7 packets)

| Opcode | Name | Description |
|--------|------|-------------|
| `0x00` | InitLS | Protocol revision (0x0106) + RSA public key |
| `0x01` | LoginServerFail | Registration rejected (reason code) |
| `0x02` | AuthResponse | Assigned serverID + serverName |
| `0x03` | PlayerAuthResponse | account + isAuthed (boolean) |
| `0x04` | KickPlayer | Force disconnect account |
| `0x05` | RequestCharacters | Request character list for account |
| `0x06` | ChangePasswordResponse | Password change result |

### 7.5 Packet Statistics

| Protocol | Direction | Regular | Extended | Total |
|----------|-----------|---------|----------|-------|
| LoginServer | Client → Server | 6 | — | **6** |
| LoginServer | Server → Client | 11 | — | **11** |
| GameServer | Client → Server | 131 | 46 | **177** |
| GameServer | Server → Client | 197 | 72 | **269** |
| GS↔LS Relay | GS → LS | 6 | — | **6** |
| GS↔LS Relay | LS → GS | 7 | — | **7** |
| **Total** | | | | **476** |

---

## 8. Binary Packet Structures

### 8.1 TCP Packet Format

Every L2 packet follows this structure:

```
+--------+--------+--------+--------+--------+---
| len_lo | len_hi | opcode |       data...
+--------+--------+--------+--------+--------+---
  byte 0   byte 1   byte 2   byte 3   ...

len = uint16 LE (total packet length INCLUDING these 2 bytes)
Minimum packet: 3 bytes (2 header + 1 opcode)
```

For **extended packets** (opcode 0xD0 for C→S, 0xFE for S→C):

```
+--------+--------+--------+--------+--------+--------+---
| len_lo | len_hi |  0xD0  | sub_lo | sub_hi |  data...
+--------+--------+--------+--------+--------+--------+---
  header (2B)      opcode    sub-opcode (2B)    payload
```

### 8.2 Init Packet (LoginServer → Client, 0x00)

| Offset | Size | Type | Field | Value | Description |
|--------|------|------|-------|-------|-------------|
| 0 | 1 | byte | opcode | `0x00` | Packet ID |
| 1 | 4 | int32 | sessionId | random | Session ID for validation |
| 5 | 4 | int32 | protocol | `0x0000C621` | Protocol version (Interlude) |
| 9 | 128 | byte[] | rsaPublicKey | scrambled | Scrambled RSA modulus (1024-bit) |
| 137 | 16 | byte[] | padding | `0x00...` | Padding zeros |
| 153 | 4 | int32 | ggConst1 | `0x29DD954E` | GameGuard constant 1 |
| 157 | 4 | int32 | ggConst2 | `0x77C39CFC` | GameGuard constant 2 |
| 161 | 4 | int32 | ggConst3 | `0x97ADB620` | GameGuard constant 3 (overflows int32!) |
| 165 | 4 | int32 | ggConst4 | `0x07BDE0F7` | GameGuard constant 4 |
| 169 | 16 | byte[] | blowfishKey | random | Dynamic Blowfish key for session |
| 185 | 1 | byte | terminator | `0x00` | End marker |

**Total payload: 186 bytes**

### 8.3 LoginOk Packet (LoginServer → Client, 0x03)

| Offset | Size | Type | Field | Value | Description |
|--------|------|------|-------|-------|-------------|
| 0 | 1 | byte | opcode | `0x03` | Packet ID |
| 1 | 4 | int32 | loginOkID1 | random | Session key part 1 |
| 5 | 4 | int32 | loginOkID2 | random | Session key part 2 |
| 9 | 4 | int32 | padding1 | `0x00` | Reserved |
| 13 | 4 | int32 | padding2 | `0x00` | Reserved |
| 17 | 4 | int32 | constant | `0x000003EA` | Unknown constant |
| 21 | 4 | int32 | padding3 | `0x00` | Reserved |
| 25 | 4 | int32 | padding4 | `0x00` | Reserved |
| 29 | 4 | int32 | padding5 | `0x00` | Reserved |
| 33 | 16 | byte[] | padding6 | `0x00...` | Reserved |

**Total payload: 49 bytes**

### 8.4 PlayOk Packet (LoginServer → Client, 0x07)

| Offset | Size | Type | Field | Value | Description |
|--------|------|------|-------|-------|-------------|
| 0 | 1 | byte | opcode | `0x07` | Packet ID |
| 1 | 4 | int32 | playOkID1 | random | Play session key part 1 |
| 5 | 4 | int32 | playOkID2 | random | Play session key part 2 |

**Total payload: 9 bytes**

### 8.5 AuthGameGuard (Client → LoginServer, 0x07)

| Offset | Size | Type | Field | Description |
|--------|------|------|-------|-------------|
| 0 | 4 | int32 | sessionId | Must match server's sessionId |
| 4 | 4 | int32 | data1 | GameGuard data (ignored) |
| 8 | 4 | int32 | data2 | GameGuard data (ignored) |
| 12 | 4 | int32 | data3 | GameGuard data (ignored) |
| 16 | 4 | int32 | data4 | GameGuard data (ignored) |

**Total payload: 20 bytes** (after opcode)

### 8.6 RequestAuthLogin (Client → LoginServer, 0x00)

| Offset | Size | Type | Field | Description |
|--------|------|------|-------|-------------|
| 0 | 128 | byte[] | raw1 | RSA-encrypted block 1 |
| 128 | 128 | byte[] | raw2 | RSA-encrypted block 2 (new method only) |

**After RSA decryption of raw1:**

| Offset | Size | Type | Field | Description |
|--------|------|------|-------|-------------|
| 0x5E | 14 | bytes | user | Login name (null-terminated ASCII) |
| 0x6C | 16 | bytes | password | Password (null-terminated ASCII) |

### 8.7 ServerList Packet (LoginServer → Client, 0x04)

**Header:**

| Offset | Size | Type | Field | Description |
|--------|------|------|-------|-------------|
| 0 | 1 | byte | opcode | `0x04` |
| 1 | 1 | byte | serverCount | Number of servers |
| 2 | 1 | byte | lastServer | Last server ID for this client |

**Per server (repeated `serverCount` times):**

| Size | Type | Field | Description |
|------|------|-------|-------------|
| 1 | byte | serverId | Server ID |
| 4 | byte[4] | ip | IP address (4 bytes) |
| 4 | int32 | port | Server port |
| 1 | byte | ageLimit | Age limit (0, 15, or 18) |
| 1 | byte | pvp | PvP flag (0 or 1) |
| 2 | int16 | currentPlayers | Current online count |
| 2 | int16 | maxPlayers | Maximum capacity |
| 1 | byte | status | 0=down, 1=up |
| 4 | int32 | serverType | Type bitmask (Normal=1, Relax=2, etc.) |
| 1 | byte | brackets | 1=show brackets before name |

**Character info section:**

| Size | Type | Field | Description |
|------|------|-------|-------------|
| 2 | int16 | unknown | `0x0000` |
| 1 | byte | charServerCount | Servers with char info |

Per server with characters:

| Size | Type | Field | Description |
|------|------|-------|-------------|
| 1 | byte | serverId | Server ID |
| 1 | byte | charCount | Characters on this account |
| 1 | byte | deleteCount | Characters pending deletion |
| 4xN | int32[] | deleteTimers | Seconds until deletion (per char) |

### 8.8 LoginFail Reason Codes

| Code | Name | Description |
|------|------|-------------|
| `0x00` | NO_MESSAGE | No message |
| `0x01` | SYSTEM_ERROR_LOGIN_LATER | System error, login later |
| `0x02` | USER_OR_PASS_WRONG | Invalid username or password |
| `0x04` | ACCESS_FAILED_TRY_LATER | Access failed, try later |
| `0x07` | ACCOUNT_IN_USE | Account already in use |
| `0x0F` | SERVER_OVERLOADED | Server overloaded |
| `0x10` | SERVER_MAINTENANCE | Server under maintenance |
| `0x14` | SYSTEM_ERROR | System error |
| `0x15` | ACCESS_FAILED | Access failed |
| `0x16` | RESTRICTED_IP | Restricted IP address |
| `0x20` | AGE_NOT_VERIFIED | Age not verified (10PM-6AM) |
| `0x23` | DUAL_BOX | Dual box detected |

### 8.9 AccountKicked Reason Codes

| Code | Name | Description |
|------|------|-------------|
| `0x01` | DATA_STEALER | Data theft detected |
| `0x08` | GENERIC_VIOLATION | Generic violation |
| `0x10` | 7_DAYS_SUSPENDED | 7-day suspension |
| `0x20` | PERMANENTLY_BANNED | Permanent ban |

---

> Previous: [02-protocol-flows.md](02-protocol-flows.md) | Next: [04-design-and-appendix.md](04-design-and-appendix.md)
