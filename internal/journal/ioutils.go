package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// SaveEntries encodes the given journal entries into JSON format and writes
// them to a file with the specified filename.
//
// The function will return an error if the encoding process fails or if the
// file cannot be created or written to.
//
// The function takes two parameters:
// - journalEntries: a slice of JournalEntry objects to be saved.
// - filename: a string representing the name of the file to write to.
//
// Example:
//
//	entries := []JournalEntry{...}
//	err := SaveEntries(entries, "journal.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func SaveEntries(journalEntries []JournalEntry, filename string) error {
	data, err := json.Marshal(Journal{Version: SchemaVersion, Entries: journalEntries})
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil

}

// LoadEntries reads the JSON file with the given filename, unmarshals its contents
// into a slice of JournalEntry, and returns the slice. If the file does not exists,
// LoadEntries creates the file and returns an empty slice. If the file cannot be
// created, LoadEntries returns an error.
func LoadEntries(filename string) ([]JournalEntry, error) {
	// Try to read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		// if the file does not exist, create it and return an empty slice
		if errors.Is(err, os.ErrNotExist) {
			// Create a new file with the current schema version
			emptyJournal := Journal{Version: SchemaVersion, Entries: []JournalEntry{}}
			jsonData, err := json.Marshal(emptyJournal)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(filename, jsonData, 0644)
			if err != nil {
				return nil, err
			}
			return []JournalEntry{}, nil
		}
		// If there was another type of error, return the error
		return nil, err
	}

	var journalData Journal
	err = json.Unmarshal(data, &journalData)
	if err != nil {
		//If we reach this section, it probably means we are using an old format file
		var oldEntries []JournalEntry

		err = json.Unmarshal(data, &oldEntries)
		if err != nil {
			return nil, err
		}
		journalData = Journal{
			Version: SchemaVersion,
			Entries: migrateEntries(oldEntries, 0),
		}
	}

	// Handle migration if necessary
	if journalData.Version < SchemaVersion {
		fmt.Printf("Detected outdated schema version %d, updating to %d...\n", journalData.Version, SchemaVersion)
		journalData.Entries = migrateEntries(journalData.Entries, journalData.Version)
		journalData.Version = SchemaVersion
		// Save the migrated data back to the file
		err = SaveEntries(journalData.Entries, filename)
		if err != nil {
			return nil, err
		}
	}

	return journalData.Entries, nil
}

func migrateEntries(entries []JournalEntry, version int) []JournalEntry {
	if version < 1 {
		// Adds default empty breaks to entries
		for i := range entries {
			if entries[i].Breaks == nil {
				entries[i].Breaks = []Break{}
			}
		}
	}
	return entries
}
