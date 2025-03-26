package journal

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"
)

func bootstrapFileContents(entries []JournalEntry, filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func TestLoadEntries(t *testing.T) {
	f, err := os.CreateTemp("", "journal_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	t.Run("When file exists and has no contents returns empty slice with no error", func(t *testing.T) {
		err := bootstrapFileContents([]JournalEntry{}, f.Name())
		if err != nil {
			t.Fatal(err)
		}
		entries, err := LoadEntries(f.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(entries) > 0 {
			t.Errorf("Expected empty slice, got: %v", entries)
		}
	})

	t.Run("When file exists with any number of entries returns populated slice with no error", func(t *testing.T) {
		entry := JournalEntry{ID: "0", StartTime: time.Now()}
		tempEntries := []JournalEntry{entry}
		err := bootstrapFileContents(tempEntries, f.Name())
		if err != nil {
			t.Fatal(err)
		}

		entries, err := LoadEntries(f.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(entries) != len(tempEntries) {
			t.Errorf("Expected populated slice %v, got: %v", tempEntries, entries)
		}
	})

	t.Run("When file does not exist we create the file and return no error", func(t *testing.T) {
		if err := os.Remove(f.Name()); err != nil {
			t.Fatal(err)
		}

		entries, err := LoadEntries(f.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected empty slice, got: %v", entries)
		}
	})

	t.Run("When file does not exist and cannot be created return nil with error", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "journal_")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dir)

		if err := os.Chmod(dir, 0444); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(dir, 0755)

		filename := dir + "/t.json"
		_, err = LoadEntries(filename)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
	})
}

func TestSaveEntries(t *testing.T) {
	// Create static time value for testing
	staticTime := time.Date(2023, time.November, 5, 12, 0, 0, 0, time.UTC)
	// Create some journal entries.
	entries := []JournalEntry{
		{
			ID:        "1",
			StartTime: staticTime,
			Notes:     generateNoteSlice(2, 0),
		},
		{
			ID:        "2",
			StartTime: staticTime,
			Notes:     generateNoteSlice(2, 2),
		},
	}

	t.Run("When saving entries file contents match specification", func(t *testing.T) {
		// Create a temporary file.
		tmpfile, err := os.CreateTemp("", "t.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name()) // clean up

		// Use the SaveEntries function to write the entries to the temporary file.
		if err := SaveEntries(entries, tmpfile.Name()); err != nil {
			t.Errorf("Error saving entries: %v\n", err)
		}

		data, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			t.Errorf("Error reading file contents: %v\n", err)
		}

		// Load the entries from the file.
		var loadedEntries Journal
		if err := json.Unmarshal(data, &loadedEntries); err != nil {
			t.Errorf("Error unmarshaling loaded entries: %v\n", err)
		}

		// Check if the loaded entries match the original entries.
		if !reflect.DeepEqual(loadedEntries.Entries, entries) {
			t.Errorf("Expected: %+v, but got: %+v\n", entries, loadedEntries)
		}
	})
}
