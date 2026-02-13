package data

import (
	"testing"
)

func TestGetMultisellEntries(t *testing.T) {
	// Load multisell data
	if err := LoadMultisell(); err != nil {
		t.Fatalf("LoadMultisell() error: %v", err)
	}

	// Skip if no multisell data available
	if len(MultisellTable) == 0 {
		t.Skip("no multisell data loaded")
	}

	// Get the first available listID
	var listID int32
	for id := range MultisellTable {
		listID = id
		break
	}

	entries := GetMultisellEntries(listID)
	if entries == nil {
		t.Fatalf("GetMultisellEntries(%d) = nil", listID)
	}

	if len(entries) == 0 {
		t.Fatalf("GetMultisellEntries(%d) returned 0 entries", listID)
	}

	// Check entry structure
	entry := entries[0]
	if entry.EntryID != 1 {
		t.Errorf("first entry ID = %d, want 1", entry.EntryID)
	}

	if len(entry.Ingredients) == 0 {
		t.Error("first entry has no ingredients")
	}
	if len(entry.Productions) == 0 {
		t.Error("first entry has no productions")
	}
}

func TestGetMultisellEntries_NotFound(t *testing.T) {
	if err := LoadMultisell(); err != nil {
		t.Fatalf("LoadMultisell() error: %v", err)
	}

	entries := GetMultisellEntries(-1)
	if entries != nil {
		t.Errorf("GetMultisellEntries(-1) = %v, want nil", entries)
	}
}

func TestFindMultisellEntry(t *testing.T) {
	if err := LoadMultisell(); err != nil {
		t.Fatalf("LoadMultisell() error: %v", err)
	}

	if len(MultisellTable) == 0 {
		t.Skip("no multisell data loaded")
	}

	var listID int32
	for id := range MultisellTable {
		listID = id
		break
	}

	// Find first entry
	entry := FindMultisellEntry(listID, 1)
	if entry == nil {
		t.Fatalf("FindMultisellEntry(%d, 1) = nil", listID)
	}
	if entry.EntryID != 1 {
		t.Errorf("entry.EntryID = %d, want 1", entry.EntryID)
	}

	// Non-existent entry
	notFound := FindMultisellEntry(listID, 99999)
	if notFound != nil {
		t.Errorf("FindMultisellEntry(%d, 99999) should be nil", listID)
	}

	// Non-existent list
	notFound = FindMultisellEntry(-1, 1)
	if notFound != nil {
		t.Error("FindMultisellEntry(-1, 1) should be nil")
	}
}
