package cmd

import (
	"fmt"
	"strconv"
	"time"

	journal "github.com/deadpyxel/workday/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// noteEditCmd represents the note edit command
var noteEditCmd = &cobra.Command{
	Use:   "edit [index] [new note]",
	Args:  cobra.ExactArgs(2),
	Short: "Edits a note in the current workday entry",
	Long: `The note edit command is used to edit a note in the current workday entry.

It requires two arguments: the index of the note to be edited and the new note text. The index must be provided as a number.
If there is no entry for the current day, the command will print an error message and return an error.
Otherwise, it will edit the note at the specified index and save the updated journal entries back to the file`,
	RunE: editNoteInCurrentDay,
}

func editNoteInCurrentDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	noteIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("Invalid index for note: %s", args[0])
	}
	newNote := args[1]

	now := time.Now()
	currenctDayId := now.Format("20060102")
	_, idx := journal.FetchEntryByID(currenctDayId, journalEntries)
	if idx == -1 {
		fmt.Println("Please run `workday start` first to create a new entry.")
		return fmt.Errorf("Could not find any entry for the current day.")
	}

	if noteIdx < 0 || noteIdx >= len(journalEntries[idx].Notes) {
		return fmt.Errorf("The index provided is not valid for the existing notes: %d", noteIdx)
	}

	journalEntries[idx].Notes[noteIdx] = journal.Note{Contents: newNote}

	err = journal.SaveEntries(journalEntries, journalPath)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully edited note %d from the current day.\n", noteIdx)
	return nil
}

func init() {
	noteCmd.AddCommand(noteEditCmd)
}
