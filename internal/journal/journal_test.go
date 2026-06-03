package journal

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func noteCompare(x, y Note) bool {
	return x.Contents == y.Contents && reflect.DeepEqual(x.Tags, y.Tags)
}

func TestAddNote(t *testing.T) {
	t.Run("when adding an empty note function returns error", func(t *testing.T) {
		entry := NewJournalEntry()
		note := Note{Contents: ""}
		err := entry.AddNote(note)
		if err == nil {
			t.Errorf("Expected zero notes with error, got %v", entry.Notes)
		}
	})
	t.Run("when adding an valid note returns no errors and appends to slice", func(t *testing.T) {
		entry := NewJournalEntry()
		note := Note{Contents: "test"}
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
		notes := []Note{{Contents: "test1"}, {Contents: "test2"}, {Contents: "test3"}}
		for _, note := range notes {
			err := entry.AddNote(note)
			if err != nil {
				t.Errorf("Expected no errors, got %v", err)
			}
		}
		if !cmp.Equal(entry.Notes, notes, cmp.Comparer(noteCompare)) {
			t.Errorf("Expected Notes field to have %v, found %v", notes, entry.Notes)
		}
	})
	t.Run("when adding a note to entry with existing notes returns no errors and updates contents", func(t *testing.T) {
		entry := NewJournalEntry()
		notes := []Note{{Contents: "test1"}, {Contents: "test2"}}
		err := entry.AddNote(notes[0])
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
		if len(entry.Notes) != 1 || !cmp.Equal(entry.Notes, []Note{notes[0]}, cmp.Comparer(noteCompare)) {
			t.Errorf("Expected to have a single note, got %v", entry.Notes)
		}
		err = entry.AddNote(notes[1])
		if err != nil {
			t.Errorf("Expected no errors, got %v", err)
		}
		if !cmp.Equal(entry.Notes, notes, cmp.Comparer(noteCompare)) {
			t.Errorf("Expected Notes field to have %v, found %v", notes, entry.Notes)
		}
	})

}

func TestJournalEntryStringer(t *testing.T) {
	startTime := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.January, 1, 13, 0, 0, 0, time.UTC)
	notes := []Note{{Contents: "Note 1"}, {Contents: "Note 2"}}

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

		expected := "Date: 2021-01-01\nStart: 12:00:00 | End: Ongoing | Time: N/A\n\n- Note 1\n- Note 2"
		result := journalEntry.String()

		if result != expected {
			t.Errorf("Expected: \n%s, but got: \n%s", expected, result)
		}
	})
}

func TestJournalEntryStringerComponents(t *testing.T) {
	startTime := time.Date(2021, time.January, 1, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2021, time.January, 1, 13, 0, 0, 0, time.UTC)
	notes := []Note{{Contents: "Note 1"}, {Contents: "Note 2"}}

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

func TestNoteStringer(t *testing.T) {
	contents := "Note 1"
	note := Note{Contents: contents}
	t.Run("When Note has no tags String returns only formatted Contents field", func(t *testing.T) {
		expected := fmt.Sprintf("- %s", contents)
		if note.String() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, note.String())
		}
	})
	t.Run("When Note has has single Tag String returns formatted Contents field and Tags field with single element", func(t *testing.T) {
		tags := []string{"Tag1"}
		note.Tags = tags
		expected := fmt.Sprintf("- %s %v", contents, tags)
		if note.String() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, note.String())
		}
	})
	t.Run("When Note has has multiple Tags String returns formatted Contents field and all tags", func(t *testing.T) {
		tags := []string{"Tag1", "Tag2"}
		note.Tags = tags
		expected := fmt.Sprintf("- %s %v", contents, tags)
		if note.String() != expected {
			t.Errorf("Expected [%s], got [%s]", expected, note.String())
		}
	})
}

func TestNoteParseContent(t *testing.T) {
	tests := []struct {
		name            string
		initialContent  string
		initialTags     []string
		expectedContent string
		expectedTags    []string
	}{
		{
			name:            "simple content without tags",
			initialContent:  "Just a simple note",
			initialTags:     nil,
			expectedContent: "Just a simple note",
			expectedTags:    nil,
		},
		{
			name:            "content with single tag",
			initialContent:  "Meeting completed #progress",
			initialTags:     nil,
			expectedContent: "Meeting completed",
			expectedTags:    []string{"progress"},
		},
		{
			name:            "content with multiple tags",
			initialContent:  "Fixed bug #bugfix #urgent",
			initialTags:     nil,
			expectedContent: "Fixed bug",
			expectedTags:    []string{"bugfix", "urgent"},
		},
		{
			name:            "content with tags and existing tags (no duplicates)",
			initialContent:  "Update done #progress #team",
			initialTags:     []string{"existing", "progress"},
			expectedContent: "Update done",
			expectedTags:    []string{"existing", "progress", "team"},
		},
		{
			name:            "content with tags and existing tags (with duplicates)",
			initialContent:  "Meeting #progress #team #urgent",
			initialTags:     []string{"team", "existing"},
			expectedContent: "Meeting",
			expectedTags:    []string{"team", "existing", "progress", "urgent"},
		},
		{
			name:            "empty content",
			initialContent:  "",
			initialTags:     []string{"existing"},
			expectedContent: "",
			expectedTags:    []string{"existing"},
		},
		{
			name:            "whitespace only content",
			initialContent:  "   ",
			initialTags:     []string{"existing"},
			expectedContent: "",
			expectedTags:    []string{"existing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := Note{
				Contents: tt.initialContent,
				Tags:     tt.initialTags,
			}

			note.ParseContent()

			if note.Contents != tt.expectedContent {
				t.Errorf("Expected content %q, got %q", tt.expectedContent, note.Contents)
			}

			if len(note.Tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(note.Tags))
				return
			}

			// Check that all expected tags are present (order may vary due to deduplication)
			tagMap := make(map[string]bool)
			for _, tag := range note.Tags {
				tagMap[tag] = true
			}

			for _, expectedTag := range tt.expectedTags {
				if !tagMap[expectedTag] {
					t.Errorf("Expected tag %q not found in result tags %v", expectedTag, note.Tags)
				}
			}
		})
	}
}

func TestNewBackfilledEntry(t *testing.T) {
	// anchor is the day the entry is backfilled for.
	anchor := time.Date(2024, 5, 27, 0, 0, 0, 0, time.UTC)
	// at builds a time-of-day on the anchor day.
	at := func(h, m int) time.Time {
		return time.Date(2024, 5, 27, h, m, 0, 0, time.UTC)
	}

	t.Run("valid minimal start and end", func(t *testing.T) {
		entry, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if entry.ID != "20240527" {
			t.Errorf("Expected ID %q, got %q", "20240527", entry.ID)
		}
		if !entry.StartTime.Equal(at(9, 0)) {
			t.Errorf("Expected StartTime %v, got %v", at(9, 0), entry.StartTime)
		}
		if !entry.EndTime.Equal(at(17, 30)) {
			t.Errorf("Expected EndTime %v, got %v", at(17, 30), entry.EndTime)
		}
		if len(entry.Breaks) != 0 {
			t.Errorf("Expected no breaks, got %v", entry.Breaks)
		}
		if len(entry.Notes) != 0 {
			t.Errorf("Expected no notes, got %v", entry.Notes)
		}
	})

	t.Run("valid with breaks inside day no overlap", func(t *testing.T) {
		breaks := []Break{
			{StartTime: at(12, 0), EndTime: at(13, 0), Reason: "lunch"},
			{StartTime: at(15, 0), EndTime: at(15, 15), Reason: "coffee"},
		}
		entry, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(entry.Breaks) != 2 {
			t.Fatalf("Expected 2 breaks, got %d", len(entry.Breaks))
		}
		if !entry.Breaks[0].StartTime.Equal(at(12, 0)) || !entry.Breaks[0].EndTime.Equal(at(13, 0)) {
			t.Errorf("Expected first break 12:00-13:00, got %v-%v", entry.Breaks[0].StartTime, entry.Breaks[0].EndTime)
		}
		if entry.Breaks[1].Reason != "coffee" {
			t.Errorf("Expected second break reason %q, got %q", "coffee", entry.Breaks[1].Reason)
		}
	})

	t.Run("valid with notes including hashtags parses tags", func(t *testing.T) {
		notes := []Note{
			{Contents: "Reviewed PRs"},
			{Contents: "Wrapped up release #progress #team"},
		}
		entry, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), nil, notes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		want := []Note{
			{Contents: "Reviewed PRs"},
			{Contents: "Wrapped up release", Tags: []string{"progress", "team"}},
		}
		if !cmp.Equal(entry.Notes, want, cmp.Comparer(noteCompare)) {
			t.Errorf("Expected notes %v, got %v", want, entry.Notes)
		}
	})

	t.Run("invalid end before start", func(t *testing.T) {
		_, err := NewBackfilledEntry(anchor, at(17, 0), at(9, 0), nil, nil)
		if err == nil {
			t.Fatalf("Expected error for end before start, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid end equal to start", func(t *testing.T) {
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(9, 0), nil, nil)
		if err == nil {
			t.Fatalf("Expected error for end equal to start, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid break starts before day start", func(t *testing.T) {
		breaks := []Break{{StartTime: at(8, 0), EndTime: at(8, 30), Reason: "early"}}
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err == nil {
			t.Fatalf("Expected error for break before day start, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid break ends after day end", func(t *testing.T) {
		breaks := []Break{{StartTime: at(17, 0), EndTime: at(18, 0), Reason: "late"}}
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err == nil {
			t.Fatalf("Expected error for break after day end, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid two breaks overlap", func(t *testing.T) {
		breaks := []Break{
			{StartTime: at(12, 0), EndTime: at(13, 0), Reason: "lunch"},
			{StartTime: at(12, 30), EndTime: at(13, 30), Reason: "overlap"},
		}
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err == nil {
			t.Fatalf("Expected error for overlapping breaks, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid break empty reason delegates to ValidateBreak", func(t *testing.T) {
		breaks := []Break{{StartTime: at(12, 0), EndTime: at(13, 0), Reason: ""}}
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err == nil {
			t.Fatalf("Expected error for empty break reason, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid break end before break start delegates to ValidateBreak", func(t *testing.T) {
		breaks := []Break{{StartTime: at(13, 0), EndTime: at(12, 0), Reason: "backwards"}}
		_, err := NewBackfilledEntry(anchor, at(9, 0), at(17, 30), breaks, nil)
		if err == nil {
			t.Fatalf("Expected error for break end before start, got nil")
		}
		if !errors.Is(err, ErrValidation) {
			t.Errorf("Expected ErrValidation, got %v", err)
		}
	})
}
