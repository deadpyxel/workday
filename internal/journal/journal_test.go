package journal

import (
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
