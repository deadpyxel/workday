package cmd

import (
	"fmt"
	"strings"
	"time"

	journal "github.com/deadpyxel/workday/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// noteCmd represents the note command
var noteCmd = &cobra.Command{
	Use:   "note [note]",
	Args:  cobra.ExactArgs(1),
	Short: "Adds a note to the current workday entry",
	Long: `The note command is used to add a note to the current workday entry.

It requires a single argument, which is the note to be added. The note must be provided as a double-quoted string.
If there is no entry for the current day, the command will print an error message and return an error.
Otherwise, it will add the note to the current entry and save the updated journal entries back to the file`,
	RunE: addNoteToCurrentDay,
}

// addNoteToCurrentDay adds a note to the current workday entry.
// It first loads the existing journal entries from the file.
// If there is no entry for the current day, it prints an error message and returns an error.
// Otherwise, it adds the note to the current entry and saves the updated journal entries back to the file.
func addNoteToCurrentDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	newNote := args[0]
	tags := strings.Split(tags, ",")
	if len(tags) == 0 {
		tags = nil
	}

	now := time.Now()
	currenctDayId := now.Format("20060102")
	_, idx := journal.FetchEntryByID(currenctDayId, journalEntries)
	if idx == -1 {
		fmt.Println("Please run `workday start` first to create a new entry.")
		return fmt.Errorf("Could not find any entry for the current day.")
	}
	journalEntries[idx].Notes = append(journalEntries[idx].Notes, journal.Note{Contents: newNote, Tags: tags})
	err = journal.SaveEntries(journalEntries, journalPath)
	if err != nil {
		return err
	}
	fmt.Println("Successfully added new note to current day.")
	return nil
}

var tags string

func init() {
	noteCmd.Flags().StringVarP(&tags, "tags", "t", "", "Tags for the note")
	rootCmd.AddCommand(noteCmd)
}
