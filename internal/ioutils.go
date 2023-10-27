package journal

import (
	"encoding/json"
	"errors"
	"os"
)

func SaveEntries(jounalEntries []JournalEntry) error {
	data, err := json.Marshal(jounalEntries)
	if err != nil {
		return err
	}
	filename := "journal.json"
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

func LoadEntries(filename string) ([]JournalEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, err := os.Create(filename)
			if err != nil {
				return nil, err
			}
			return []JournalEntry{}, nil
		}
		return nil, err
	}

	var entries []JournalEntry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
