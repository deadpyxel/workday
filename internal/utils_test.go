package journal

import (
	"strconv"
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

func TestFetchEntriesByWeek(t *testing.T) {
	currentDate := time.Now()
	currYear, currWeek := currentDate.ISOWeek()

	entries := []JournalEntry{
		{ID: "1", StartTime: time.Now().AddDate(0, 0, -7), Notes: []string{"Note 1", "Note 2"}}, // Entry from one week ago
		{ID: "2", StartTime: time.Now(), Notes: []string{"Note 3", "Note 4"}},                   // Entry for today
		{ID: "3", StartTime: time.Now().AddDate(0, 0, 7), Notes: []string{"Note 5", "Note 6"}},  // Entry for next week
	}

	testCases := []struct {
		name           string
		entries        []JournalEntry
		expectedErr    string
		expectedResult []JournalEntry
	}{
		{name: "When passing an empty slice function returns an error", entries: []JournalEntry{}, expectedErr: "No entries were passed", expectedResult: nil},
		{name: "When passing a slice with no date in current week returns an error", entries: entries[:1], expectedErr: "No entries found for the current week", expectedResult: nil},
		{name: "When passing a slice with entries in current week returns filtered slice with no errors", entries: entries, expectedErr: "", expectedResult: []JournalEntry{entries[1]}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FetchEntriesByWeekDate(tc.entries, currentDate)
			// If we expcted a nil result, but a non nil error
			if tc.expectedResult == nil && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			// If the err is not nil, but the message does not match
			if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("Expected error: %s, but got %s", tc.expectedErr, err.Error())
			}
			// If the result is not nil, context should match expectation
			if result != nil {
				resultYear, resultWeek := result[0].StartTime.ISOWeek()
				if len(result) != len(tc.expectedResult) {
					t.Errorf("Expected result to be %v, got %v", tc.expectedResult, result)
				}
				if resultYear != currYear || resultWeek != currWeek {
					t.Errorf("The resulting entries are not from the current week")
				}
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
