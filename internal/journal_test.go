package journal

import "testing"

func TestFindEntryByID(t *testing.T) {
	entries := []JournalEntry{}

	entry := FetchEntryByID("0000", entries)
	if entry == nil {
		t.Errorf("Expected an entry, got %v", entry)
	}
}
