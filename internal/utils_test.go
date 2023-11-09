package journal

import (
	"strconv"
	"testing"
)

func TestFindEntryByID(t *testing.T) {
	entryID := "0000"
	entries := []JournalEntry{{ID: entryID}}
	expectedEntry := entries[0]
	testCases := []struct {
		name        string
		entries     []JournalEntry
		expectedId  string
		expectedIdx int
		expected    *JournalEntry
	}{
		{
			name:        "When entry exists should return it",
			entries:     entries,
			expectedId:  entryID,
			expectedIdx: 0,
			expected:    &expectedEntry,
		}, {
			name:        "When entry does not exists should return nil",
			entries:     entries,
			expectedId:  "9999",
			expectedIdx: -1,
			expected:    nil,
		}, {
			name:        "When slice is empty should return nil",
			entries:     []JournalEntry{},
			expectedId:  "1234",
			expectedIdx: -1,
			expected:    nil,
		}, {
			name:        "When multiple entries returns correct entry",
			entries:     append(entries, JournalEntry{ID: "1111"}),
			expectedId:  "0000",
			expectedIdx: 0,
			expected:    &expectedEntry,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry, idx := FetchEntryByID(tc.expectedId, tc.entries)
			if entry != nil && tc.expected != nil {
				if entry.ID != tc.expected.ID || idx != tc.expectedIdx {
					t.Errorf("Expected (%v, idx: %d), got (%v, idx: %d)", tc.expected, tc.expectedIdx, entry, idx)
				}
			} else if entry != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, entry)
			}
		})
	}
}
func BenchmarkFetchEntryByID(b *testing.B) {
	// Setup entries for benchmark
	entries := make([]JournalEntry, 1e6)
	for i := range entries {
		entries[i] = JournalEntry{ID: strconv.Itoa(i)}
	}

	// Run the Benchmark
	for i := 0; i < b.N; i++ {
		FetchEntryByID(strconv.Itoa(i%1e6), entries)
	}
}
