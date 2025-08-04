package journal

import (
	"fmt"
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

func generateNoteSlice(size int, start int) []Note {
	var notes []Note
	for i := start; i < size; i++ {
		notes = append(notes, Note{Contents: fmt.Sprintf("Note %d", i)})
	}
	return notes
}

func TestFetchEntriesByWeek(t *testing.T) {
	currentDate := time.Now()
	currYear, currWeek := currentDate.ISOWeek()

	entries := []JournalEntry{
		{ID: "1", StartTime: time.Now().AddDate(0, 0, -7), Notes: generateNoteSlice(2, 0)}, // Entry from one week ago
		{ID: "2", StartTime: time.Now(), Notes: generateNoteSlice(2, 2)},                   // Entry for today
		{ID: "3", StartTime: time.Now().AddDate(0, 0, 7), Notes: generateNoteSlice(2, 4)},  // Entry for next week
	}

	testCases := []struct {
		name           string
		entries        []JournalEntry
		expectedErr    string
		expectedResult []JournalEntry
	}{
		{name: "When passing an empty slice function returns an error", entries: []JournalEntry{}, expectedErr: "no entries found: no entries found for weekly report", expectedResult: nil},
		{name: "When passing a slice with no date in current week returns an error", entries: entries[:1], expectedErr: "no entries found: no entries found for current week", expectedResult: nil},
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

func TestCalculateTotalTime(t *testing.T) {
	t.Run("When slice is empty returns 0 with no errors", func(t *testing.T) {
		entries := []JournalEntry{}
		expected := 0 * time.Second
		result, err := CalculateTotalTime(entries)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != expected {
			t.Errorf("Expected %v, but got %v", expected, result)
		}
	})
	t.Run("When slice contains valid entries, returns expected result with no errors", func(t *testing.T) {
		entries := []JournalEntry{
			{StartTime: time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC), EndTime: time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)},
			{StartTime: time.Date(2021, time.January, 1, 14, 0, 0, 0, time.UTC), EndTime: time.Date(2021, time.January, 1, 16, 0, 0, 0, time.UTC)},
		}
		expected := 4 * time.Hour
		result, err := CalculateTotalTime(entries)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != expected {
			t.Errorf("Expected %v, but got %v", expected, result)
		}
	})

	t.Run("When slice contains invalid entries, returns 0 with error", func(t *testing.T) {
		entries := []JournalEntry{
			{StartTime: time.Date(2021, time.January, 1, 10, 0, 0, 0, time.UTC), EndTime: time.Date(2021, time.January, 1, 9, 0, 0, 0, time.UTC)},
		}
		expected := 0 * time.Second
		result, err := CalculateTotalTime(entries)
		if err == nil {
			t.Errorf("Expected an error, got nil")
		}
		if result != expected {
			t.Errorf("Expected %v, but got %v", expected, result)
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

func TestParseNoteTags(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedTags    []string
	}{
		{
			name:            "simple note without tags",
			input:           "Just a simple note",
			expectedContent: "Just a simple note",
			expectedTags:    nil,
		},
		{
			name:            "note with single tag",
			input:           "Meeting completed #progress",
			expectedContent: "Meeting completed",
			expectedTags:    []string{"progress"},
		},
		{
			name:            "note with multiple tags",
			input:           "Fixed bug #bugfix #urgent #team-alpha",
			expectedContent: "Fixed bug",
			expectedTags:    []string{"bugfix", "urgent", "team-alpha"},
		},
		{
			name:            "note with tags in middle",
			input:           "Started #project review and finished #testing",
			expectedContent: "Started review and finished",
			expectedTags:    []string{"project", "testing"},
		},
		{
			name:            "note with tags and extra whitespace",
			input:           "Meeting   done    #progress   #team   ",
			expectedContent: "Meeting done",
			expectedTags:    []string{"progress", "team"},
		},
		{
			name:            "empty string",
			input:           "",
			expectedContent: "",
			expectedTags:    nil,
		},
		{
			name:            "whitespace only",
			input:           "   ",
			expectedContent: "",
			expectedTags:    nil,
		},
		{
			name:            "tags with underscores and hyphens",
			input:           "Update #bug_fix #team-sync #v1_2_3",
			expectedContent: "Update",
			expectedTags:    []string{"bug_fix", "team-sync", "v1_2_3"},
		},
		{
			name:            "tags with numbers",
			input:           "Issue resolved #bug123 #sprint2024",
			expectedContent: "Issue resolved",
			expectedTags:    []string{"bug123", "sprint2024"},
		},
		{
			name:            "invalid tags (special characters)",
			input:           "Test #invalid@tag #valid-tag #another$invalid",
			expectedContent: "Test @tag $invalid",
			expectedTags:    []string{"invalid", "valid-tag", "another"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, tags := ParseNoteTags(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
			}

			if len(tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(tags))
				return
			}

			for i, expectedTag := range tt.expectedTags {
				if i >= len(tags) || tags[i] != expectedTag {
					t.Errorf("Expected tag[%d] to be %q, got %q", i, expectedTag, tags[i])
				}
			}
		})
	}
}

func TestFormatNoteWithTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		tags     []string
		expected string
	}{
		{
			name:     "note without tags",
			content:  "Simple note",
			tags:     nil,
			expected: "Simple note",
		},
		{
			name:     "note with empty tags",
			content:  "Simple note",
			tags:     []string{},
			expected: "Simple note",
		},
		{
			name:     "note with single tag",
			content:  "Meeting done",
			tags:     []string{"progress"},
			expected: "Meeting done #progress",
		},
		{
			name:     "note with multiple tags",
			content:  "Bug fixed",
			tags:     []string{"bugfix", "urgent", "team-alpha"},
			expected: "Bug fixed #bugfix #urgent #team-alpha",
		},
		{
			name:     "empty content with tags",
			content:  "",
			tags:     []string{"tag1", "tag2"},
			expected: " #tag1 #tag2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatNoteWithTags(tt.content, tt.tags)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
