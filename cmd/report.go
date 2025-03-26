package cmd

import (
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Reports the workday entry for the current day",
	Long: `The report command is used to print out the workday entry for the current day.

It loads the existing jorunal entries from the file and fetches the entry for the current day.
If there is no entry for the current day, it returns an error.
Otherwise, it prints out the entry.`,
	RunE: reportWorkDay,
}

// reportWorkDay reports the workday entry for the current day.
// It first loads the existing journal entries from the file.
// If there is no entry for the current day, it returns and error.
// Otherwise, it prints out the entry.
func reportWorkDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}
	now := time.Now()
	currenctDayId := now.Format("20060102")
	currentEntry, _ := journal.FetchEntryByID(currenctDayId, journalEntries)
	if currentEntry == nil {
		return fmt.Errorf("Could not find any entry for the current day.")
	}
	fmt.Println(currentEntry) // Print basic work info

	// Prints break information
	if len(currentEntry.Breaks) > 0 {
		fmt.Printf("\nBreaks:\n")
		for _, br := range currentEntry.Breaks {
			startTime := br.StartTime.Format("15:04:05")
			endTime := "Ongoing"
			if !br.EndTime.IsZero() {
				endTime = br.EndTime.Format("15:04:05")
			}
			fmt.Printf("\t- Start: %s, End: %s, Reason: %s\n", startTime, endTime, br.Reason)
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
