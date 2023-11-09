package journal

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestAddNote(t *testing.T) {
	t.Run("when adding an empty note function returns error", func(t *testing.T) {
		entry := NewJournalEntry()
		note := ""
		err := entry.AddNote(note)
		if err == nil {
			t.Errorf("Expected zero notes with error, got %v", entry.Notes)
		}
	})
	t.Run("when adding an valid note returns no errors and appends to slice", func(t *testing.T) {
		entry := NewJournalEntry()
		note := "test"
		err := entry.AddNote(note)
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
		if len(entry.Notes) != 1 {
			t.Errorf("Expected Notes field to have a single note, found %v", entry.Notes)
		}
	})
	t.Run("when adding multiple notes returns no errors and has all contents", func(t *testing.T) {
		entry := NewJournalEntry()
		notes := []string{"test1", "test2", "test3"}
		for _, note := range notes {
			err := entry.AddNote(note)
			if err != nil {
				t.Errorf("Expected no errors, got %v", err)
			}
		}
		if slices.Compare(entry.Notes, notes) != 0 {
			t.Errorf("Expected Notes field to have %v, found %v", notes, entry.Notes)
		}
	})
	t.Run("when adding a note to entry with existing notes returns no errors and updates contents", func(t *testing.T) {
		entry := NewJournalEntry()
		notes := []string{"test1", "test2"}
		err := entry.AddNote(notes[0])
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
		if len(entry.Notes) != 1 || slices.Compare(entry.Notes, []string{notes[0]}) != 0 {
			t.Errorf("Expected to have a single note, got %v", entry.Notes)
		}
		err = entry.AddNote(notes[1])
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
		if slices.Compare(entry.Notes, notes) != 0 {
			t.Errorf("Expected Notes field to have %v, found %v", notes, entry.Notes)
		}
	})

}

func TestJournalEntryStringer(t *testing.T) {
	startTime := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.January, 1, 13, 0, 0, 0, time.UTC)
	notes := []string{"Note 1", "Note 2"}

	t.Run("When an entry has all fields filled returns formatted string", func(t *testing.T) {
		journalEntry := &JournalEntry{
			StartTime: startTime,
			EndTime:   endTime,
			Notes:     notes,
		}

		expected := "Date: 2021-01-01\nStart: 12:00:00 | End: 13:00:00 | Time: 1h0m0s\n\n- Note 1\n- Note 2"
		result := journalEntry.String()

		if result != expected {
			t.Errorf("Expected: \n%s, but got: \n%s", expected, result)
		}
	})
	t.Run("When an entry has no EndTime returns formatted string with not closed message", func(t *testing.T) {
		journalEntry := &JournalEntry{
			StartTime: startTime,
			Notes:     notes,
		}

		expected := "Date: 2021-01-01\nStart: 12:00:00 | End: Not yet closed | Time: N/A\n\n- Note 1\n- Note 2"
		result := journalEntry.String()

		if result != expected {
			t.Errorf("Expected: \n%s, but got: \n%s", expected, result)
		}
	})
}

func TestJournalEntryStringerComponents(t *testing.T) {
	startTime := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.January, 1, 13, 0, 0, 0, time.UTC)
	notes := []string{"Note 1", "Note 2"}

	journalEntry := &JournalEntry{
		StartTime: startTime,
		EndTime:   endTime,
		Notes:     notes,
	}

	expectedHeader := "Date: 2021-01-01"
	expectedTime := "Start: 12:00:00 | End: 13:00:00 | Time: 1h0m0s"
	expectedNotes := []string{"- Note 1", "- Note 2"}

	result := journalEntry.String()

	// Split the result into lines
	lines := strings.Split(strings.TrimSpace(result), "\n")

	t.Run("Date header is properly formatted", func(t *testing.T) {
		if lines[0] != expectedHeader {
			t.Errorf("Expected header: %s, but got: %s", expectedHeader, lines[0])
		}
	})

	t.Run("Time header is properly formatted", func(t *testing.T) {
		if lines[1] != expectedTime {
			t.Errorf("Expected time: %s, but got: %s", expectedTime, lines[1])
		}
	})

	t.Run("Notes components is properly formatted", func(t *testing.T) {
		for i, note := range expectedNotes {
			// Add 3 because the notes start on the 4th line
			if lines[i+3] != note {
				t.Errorf("Expected note: %s, but got: %s", note, lines[i+3])
			}
		}
	})
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
