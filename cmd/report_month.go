package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reportMonthCmd represents the report command for generating a report for the current month.
var reportMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Generates a report for the current month",
	Long: `The month command generates a report for the current month.
It loads the existing journal entries from the file and fetches the entries for the current month.
If there are no entries for the current month, it returns an error.
Otherwise, it prints out the entries.`,
	RunE: reportMonth,
}

// reportMonth loads the existing journal entries from the file and reports the entries for the current month.
// It first retrieves the journal path from the viper configuration and loads the journal entries from this path.
// If there is an error while loading the entries, it returns the error.
// It then gets the current time and fetches the entries for the current month.
// If there is an error while fetching the entries, it returns the error.
// It then iterates over the entries for the current month and prints each entry.
// It formats the current month and year and calculates the total time for the current month.
// It then prints the number of entries found for the current month, the formatted month and year, and the total time.
// It returns nil if there are no errors.
func reportMonth(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Check if the month flag has been set
	monthFlag, _ := cmd.Flags().GetString("month")
	var monthFilter time.Time
	if monthFlag != "" {
		monthFilter, err = time.Parse("2006-01", monthFlag)
		if err != nil {
			return errors.New("invalid month format, expected YYYY-MM")
		}
	} else {
		monthFilter = time.Now()
	}
	currentMonth, err := journal.FetchEntriesByMonthDate(entries, monthFilter)
	if err != nil {
		return err
	}
	for _, entry := range currentMonth {
		fmt.Printf("%s\n---\n", entry.String())
	}
	month := monthFilter.Format("January 2006")
	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return err
	}
	totalLunchTime := lunchTime * time.Duration(len(currentMonth))
	totalTime, err := journal.CalculateTotalTime(currentMonth)
	if err != nil {
		return err
	}
	totalTime -= totalLunchTime
	fmt.Printf("> %d entries found for %s, totalling %v of work...\n", len(currentMonth), month, totalTime)
	return nil
}

func init() {
	reportMonthCmd.Flags().StringP("month", "m", "", "Specify the month and year in the format YYYY-MM")
	reportCmd.AddCommand(reportMonthCmd)
}
