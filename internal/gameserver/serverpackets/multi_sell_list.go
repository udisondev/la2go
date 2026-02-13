package serverpackets

import (
	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMultiSellList is the opcode for MultiSellList packet (S2C 0xD0).
// Sends multisell dialog with ingredient → product exchange entries.
//
// Phase 8: Trade System Foundation.
// Java reference: MultiSellList.java
const OpcodeMultiSellList = 0xD0

// multiSellPageSize is the maximum number of entries per page.
const multiSellPageSize = 40

// MultiSellList packet (S2C 0xD0) shows multisell exchange dialog.
//
// Packet structure:
//   - opcode (byte) — 0xD0
//   - listId (int32)
//   - page (int32) — 1-based page number
//   - finished (int32) — 1 if last page, 0 otherwise
//   - pageSize (int32) — always 40
//   - size (int32) — entries on this page
//   - for each entry:
//   - entryId (int32)
//   - unknown1 (int32) — 0
//   - unknown2 (int32) — 0
//   - unknown3 (byte) — 1
//   - productCount (short)
//   - ingredientCount (short)
//   - for each product:
//   - displayId (short) — item template ID
//   - bodyPart (int32) — 0 for non-equipment
//   - type2 (short) — 0xFFFF for special currency
//   - itemCount (int32)
//   - enchantLevel (short) — 0
//   - augmentId (int32) — 0
//   - mana (int32) — 0
//   - for each ingredient:
//   - displayId (short) — item template ID
//   - type2 (short) — 0xFFFF for special currency
//   - itemCount (int32)
//   - enchantLevel (short) — 0
//   - augmentId (int32) — 0
//   - mana (int32) — 0
//
// Phase 8: Trade System Foundation.
type MultiSellList struct {
	ListID  int32
	Entries []data.MultisellEntry
	Page    int32 // 1-based, 0 means auto-paginate from page 1
}

// NewMultiSellList creates MultiSellList packet for a given page.
// If page is 0, sends all entries as page 1.
func NewMultiSellList(listID int32, entries []data.MultisellEntry, page int32) *MultiSellList {
	if page <= 0 {
		page = 1
	}
	return &MultiSellList{
		ListID:  listID,
		Entries: entries,
		Page:    page,
	}
}

// Write serializes MultiSellList packet to bytes.
func (p *MultiSellList) Write() ([]byte, error) {
	// Calculate page slice
	startIdx := (int(p.Page) - 1) * multiSellPageSize
	if startIdx >= len(p.Entries) {
		startIdx = 0
	}
	endIdx := min(startIdx+multiSellPageSize, len(p.Entries))
	pageEntries := p.Entries[startIdx:endIdx]

	totalPages := (len(p.Entries) + multiSellPageSize - 1) / multiSellPageSize
	finished := int32(0)
	if int(p.Page) >= totalPages {
		finished = 1
	}

	// Estimate size: header(21) + entries * (17 + products*22 + ingredients*18)
	estimatedSize := 21
	for _, entry := range pageEntries {
		estimatedSize += 17 + len(entry.Productions)*22 + len(entry.Ingredients)*18
	}

	w := packet.NewWriter(estimatedSize)

	w.WriteByte(OpcodeMultiSellList)
	w.WriteInt(p.ListID)
	w.WriteInt(p.Page)
	w.WriteInt(finished)
	w.WriteInt(int32(multiSellPageSize))
	w.WriteInt(int32(len(pageEntries)))

	for _, entry := range pageEntries {
		w.WriteInt(entry.EntryID) // entryId
		w.WriteInt(0)             // unknown1 (C6)
		w.WriteInt(0)             // unknown2 (C6)
		w.WriteByte(1)            // unknown3

		w.WriteShort(int16(len(entry.Productions)))
		w.WriteShort(int16(len(entry.Ingredients)))

		// Products
		for _, prod := range entry.Productions {
			w.WriteShort(int16(prod.ItemID)) // displayId
			w.WriteInt(0)                    // bodyPart (0 for non-equipment)
			w.WriteShort(0)                  // type2
			w.WriteInt(int32(prod.Count))    // itemCount
			w.WriteShort(0)                  // enchantLevel
			w.WriteInt(0)                    // augmentId
			w.WriteInt(0)                    // mana
		}

		// Ingredients
		for _, ing := range entry.Ingredients {
			w.WriteShort(int16(ing.ItemID)) // displayId
			w.WriteShort(0)                 // type2
			w.WriteInt(int32(ing.Count))    // itemCount
			w.WriteShort(0)                 // enchantLevel
			w.WriteInt(0)                   // augmentId
			w.WriteInt(0)                   // mana
		}
	}

	return w.Bytes(), nil
}
