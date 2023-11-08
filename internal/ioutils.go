package journal

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/viper"
)

func SaveEntries(jounalEntries []JournalEntry) error {
	data, err := json.Marshal(jounalEntries)
	if err != nil {
		return err
	}
	filename := viper.GetString("journalPath")
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
			_, err := os.Create(filename)
			if err != nil {
				return nil, err
			}
			return []JournalEntry{}, nil
		}
		// If there was another type of error, return the error
		return nil, err
	}

	// Unmarshal JSON data into a slice of JournalEntry
	var entries []JournalEntry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
