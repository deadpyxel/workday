/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	journal "github.com/deadpyxel/workday/internal"
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
	currentWeek, err := journal.CurrentWeekEntries(journalEntries)
	if err != nil {
		return err
	}
	for _, entry := range currentWeek {
		fmt.Printf("%s---\n", entry.String())
	}
	return nil
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
