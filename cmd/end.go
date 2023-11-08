/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	journal "github.com/deadpyxel/workday/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// endCmd represents the end command
var endCmd = &cobra.Command{
	Use:   "end",
	Short: "Marks the current workday as finished",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: markDayAsFinished,
}

// markDayAsFinished marks the current day's JournalEntry as finished.
// It loads the journal entries from the file, finds the entry for the current day,
// and sets its EndTime to the current time.
// If no entry is found for the current day, it returns and error.
// After modifying the entry, it saves the updated entries back to the file.
func markDayAsFinished(cmd *cobra.Command, args []string) error {
	// Get current date
	now := time.Now()
	currentDayId := now.Format("20060102")

	// Load journal entries
	entries, err := journal.LoadEntries(viper.GetString("journalPath"))
	if err != nil {
		return err
	}

	// Get the index of the entry on the slice, -1 if not found
	_, entryIndex := journal.FetchEntryByID(currentDayId, entries)

	if entryIndex == -1 {
		return fmt.Errorf("No entry found for the current day")
	}

	entries[entryIndex].EndDay()

	err = journal.SaveEntries(entries)
	if err != nil {
		return fmt.Errorf("Failed to save journal entries: %v\n", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(endCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// endCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// endCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
