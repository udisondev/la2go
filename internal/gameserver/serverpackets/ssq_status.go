package serverpackets

import (
	"math"

	"github.com/udisondev/la2go/internal/game/sevensigns"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSSQStatus is the opcode for the Seven Signs status packet (S2C 0xF5).
//
// Java reference: SSQStatus.java.
const OpcodeSSQStatus byte = 0xF5

// SSQStatus sends Seven Signs status to the client.
//
// Packet structure (S2C 0xF5):
//   - opcode          byte   0xF5
//   - page            byte   1-4
//   - currentPeriod   byte   0-3
//
// Page 1: Player info, overall scores.
// Page 2: Festival scores.
// Page 3: Seal proportions.
// Page 4: Seal predictions.
type SSQStatus struct {
	Page          byte
	CurrentPeriod sevensigns.Period
	CurrentCycle  int32

	// Player data (page 1).
	PlayerCabal     sevensigns.Cabal
	PlayerSeal      sevensigns.Seal
	PlayerStones    int32 // total contribution score
	PlayerAdena     int32 // ancient adena collected

	// Scores (page 1).
	DawnStoneScore  float64
	DuskStoneScore  float64
	DawnFestival    int32
	DuskFestival    int32

	// Seal data (page 3).
	SealOwners      [4]sevensigns.Cabal // index 1-3
	DawnSealMembers [4]int32            // per-seal Dawn member count
	DuskSealMembers [4]int32            // per-seal Dusk member count
	TotalDawnMembers int32
	TotalDuskMembers int32

	// Predictions (page 4).
	WinnerCabal sevensigns.Cabal
}

// Write serializes the SSQStatus packet.
func (p *SSQStatus) Write() ([]byte, error) {
	w := packet.NewWriter(256)
	w.WriteByte(OpcodeSSQStatus)
	w.WriteByte(p.Page)
	w.WriteByte(byte(p.CurrentPeriod))

	switch p.Page {
	case 1:
		p.writePage1(w)
	case 2:
		p.writePage2(w)
	case 3:
		p.writePage3(w)
	case 4:
		p.writePage4(w)
	}

	return w.Bytes(), nil
}

func (p *SSQStatus) writePage1(w *packet.Writer) {
	w.WriteInt(p.CurrentCycle)

	// Описание периода (SystemMessageId).
	w.WriteInt(periodDescMsgID(p.CurrentPeriod))
	w.WriteInt(periodEndMsgID(p.CurrentPeriod))

	w.WriteByte(byte(p.PlayerCabal))
	w.WriteByte(byte(p.PlayerSeal))
	w.WriteInt(p.PlayerStones)
	w.WriteInt(p.PlayerAdena)

	totalStone := p.DawnStoneScore + p.DuskStoneScore

	var dawnStoneProp, duskStoneProp int32
	if totalStone > 0 {
		duskStoneProp = int32(math.Round(p.DuskStoneScore / totalStone * 500))
		dawnStoneProp = int32(math.Round(p.DawnStoneScore / totalStone * 500))
	}

	dawnTotal := dawnStoneProp + p.DawnFestival
	duskTotal := duskStoneProp + p.DuskFestival
	totalOverall := dawnTotal + duskTotal

	var dawnPct, duskPct byte
	if totalOverall > 0 {
		dawnPct = byte(math.Round(float64(dawnTotal) / float64(totalOverall) * 100))
		duskPct = byte(math.Round(float64(duskTotal) / float64(totalOverall) * 100))
	}

	// Dusk scores.
	w.WriteInt(duskStoneProp)
	w.WriteInt(p.DuskFestival)
	w.WriteInt(duskTotal)
	w.WriteByte(duskPct)

	// Dawn scores.
	w.WriteInt(dawnStoneProp)
	w.WriteInt(p.DawnFestival)
	w.WriteInt(dawnTotal)
	w.WriteByte(dawnPct)
}

func (p *SSQStatus) writePage2(w *packet.Writer) {
	w.WriteShort(1)                     // unknown constant
	w.WriteByte(sevensigns.FestivalCount) // festival count

	// В MVP без реальных festival данных — заполняем нулями.
	for i := range sevensigns.FestivalCount {
		w.WriteByte(byte(i + 1))
		w.WriteInt(int32(festivalMaxScore(i)))

		// Dusk: score=0, members=0.
		w.WriteInt(0)
		w.WriteByte(0)

		// Dawn: score=0, members=0.
		w.WriteInt(0)
		w.WriteByte(0)
	}
}

func (p *SSQStatus) writePage3(w *packet.Writer) {
	w.WriteByte(10) // min % to retain
	w.WriteByte(35) // min % to claim
	w.WriteByte(3)  // seal count

	for i := int32(1); i <= 3; i++ {
		seal := sevensigns.Seal(i)
		w.WriteByte(byte(seal))
		w.WriteByte(byte(p.SealOwners[i]))

		var duskPct, dawnPct byte
		totalMembers := p.TotalDawnMembers + p.TotalDuskMembers
		if totalMembers > 0 {
			duskPct = byte(float64(p.DuskSealMembers[i]) / float64(totalMembers) * 100)
			dawnPct = byte(float64(p.DawnSealMembers[i]) / float64(totalMembers) * 100)
		}
		w.WriteByte(duskPct)
		w.WriteByte(dawnPct)
	}
}

func (p *SSQStatus) writePage4(w *packet.Writer) {
	w.WriteByte(byte(p.WinnerCabal))
	w.WriteByte(3) // seal count

	for i := int32(1); i <= 3; i++ {
		w.WriteByte(byte(i))
		w.WriteByte(byte(p.SealOwners[i]))
		w.WriteShort(0) // outcome system message (simplified)
	}
}

// periodDescMsgID maps a period to its system message description ID.
func periodDescMsgID(p sevensigns.Period) int32 {
	switch p {
	case sevensigns.PeriodRecruitment:
		return 1255 // "This is the initial period..."
	case sevensigns.PeriodCompetition:
		return 1256 // "This is the quest event period..."
	case sevensigns.PeriodResults:
		return 1257 // "This is the seal validation period..."
	case sevensigns.PeriodSealValidation:
		return 1258 // "During the seal validation period..."
	default:
		return 0
	}
}

// periodEndMsgID maps a period to its "period ends..." system message ID.
func periodEndMsgID(p sevensigns.Period) int32 {
	switch p {
	case sevensigns.PeriodRecruitment:
		return 1259
	case sevensigns.PeriodCompetition:
		return 1260
	case sevensigns.PeriodResults:
		return 1261
	case sevensigns.PeriodSealValidation:
		return 1262
	default:
		return 0
	}
}

// festivalMaxScore returns the max possible score for a festival tier.
func festivalMaxScore(tier int) int32 {
	scores := [sevensigns.FestivalCount]int32{60, 70, 100, 120, 150}
	if tier < 0 || tier >= len(scores) {
		return 0
	}
	return scores[tier]
}
