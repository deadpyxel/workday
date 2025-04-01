package cmd

import (
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reportDate string

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

	var tgtEntry *journal.JournalEntry
	// If no date is provided, report the current day
	tgtDay := time.Now()

	if reportDate != "" {
		// if a date is provided, try to parse it and find the correponding entry
		tgtDay, err = time.Parse("2006-01-02", reportDate)
		if err != nil {
			return fmt.Errorf("invalid date format, Use YYYY-MM-DD: %v", err)
		}
	}

	tgtDayID := tgtDay.Format("20060102")
	tgtEntry, _ = journal.FetchEntryByID(tgtDayID, journalEntries)
	if tgtEntry == nil {
		return fmt.Errorf("Could not find any entry for the date: %s", reportDate)
	}
	fmt.Println(tgtEntry) // Print basic work info

	// Prints break information
	if len(tgtEntry.Breaks) > 0 {
		fmt.Printf("\nBreaks:\n")
		for _, br := range tgtEntry.Breaks {
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

	// Specify date for report
	reportCmd.Flags().StringVarP(&reportDate, "date", "d", "", "Specify the date in YYYY-MM-DD format")
}
