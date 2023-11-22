package cmd

import (
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportCmd represents the report command
var reportWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Generates a report for the current week",
	Long: `The week command generates a report for the current week.
It loads the existing journal entries from the file and fetches the entries for the current week.
If there are no entries for the current week, it returns an error.
Otherwise, it prints out the entries.`,
	RunE: reportWeek,
}

// reportWorkDay reports the workday entry for the current day.
// It first loads the existing journal entries from the file.
// If there is no entry for the current day, it returns and error.
// Otherwise, it prints out the entry.
func reportWeek(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	journalEntries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}
	now := time.Now()
	currentWeek, err := journal.FetchEntriesByWeekDate(journalEntries, now)
	if err != nil {
		return err
	}
	for _, entry := range currentWeek {
		fmt.Printf("%s\n---\n", entry.String())
	}
	return nil
}

func init() {
	reportCmd.AddCommand(reportWeekCmd)
}
