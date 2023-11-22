package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type TempJournalEntry struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Notes     []string
}

func loadTempEntries(journalPath string) ([]TempJournalEntry, error) {
	file, err := os.Open(journalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tempEntries []TempJournalEntry
	err = json.NewDecoder(file).Decode(&tempEntries)
	if err != nil {
		return nil, err
	}

	return tempEntries, nil
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrates an existing journal.json to the latest data storage format.",
	RunE:  migrateJournal,
}

func migrateJournal(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	tempEntries, err := loadTempEntries(journalPath)
	if err != nil {
		return err
	}

	journalEntries := make([]journal.JournalEntry, len(tempEntries))
	for i, tempEntry := range tempEntries {
		journalEntries[i] = journal.JournalEntry{
			ID:        tempEntry.ID,
			StartTime: tempEntry.StartTime,
			EndTime:   tempEntry.EndTime,
			Notes:     make([]journal.Note, len(tempEntry.Notes)),
		}
		for j, tempNote := range tempEntry.Notes {
			journalEntries[i].Notes[j] = journal.Note{
				Contents: tempNote,
			}
		}
	}

	err = journal.SaveEntries(journalEntries, journalPath)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully migrated %s to latest format\n", journalPath)
	return nil
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
