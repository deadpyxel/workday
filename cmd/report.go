/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	journal "github.com/deadpyxel/workday/internal"
	"github.com/spf13/cobra"
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
	journalEntries, err := journal.LoadEntries("journal.json")
	if err != nil {
		return err
	}
	now := time.Now()
	currenctDayId := now.Format("20060102")
	currentEntry, _ := journal.FetchEntryByID(currenctDayId, journalEntries)
	if currentEntry == nil {
		return fmt.Errorf("Could not find any entry for the current day.")
	}
	fmt.Println(currentEntry)
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
