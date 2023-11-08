/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	journal "github.com/deadpyxel/workday/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a new workday entry",
	Long: `The start command is used to begin a new workday entry in the journal.

It creates a new JournalEntry with the current date and time as the start time,
appends it to the existing journal entries, and saves the updated journal entries to the file.
After running this command, you can begin adding notes to the new workday entry.`,
	RunE: startWorkDay,
}

// startWorkDay starts a new workday entry in the journal.
// It first loads the existing journal entries from the file.
// If there is already an entry for the current day, it asks the user if they want to override it.
// If the user agrees, it overwrites the existing entry with a new one.
// If the user does not agree, it does nothing and returns nil.
// If there is no entry for the current day, it creates a new JournalEntry with the current date and time as the start time,
// appends the new entry to the journal entries, and saves the updated journal entries back to the file.
// It then prints a message indicating that a new JournalEntry has been added for the current day.
func startWorkDay(cmd *cobra.Command, args []string) error {
	journalEntries, err := journal.LoadEntries(viper.GetString("journalPath"))
	if err != nil {
		return err
	}

	now := time.Now()
	currenctDayId := now.Format("20060102")
	dateStr := now.Format("2006-01-02")
	_, idx := journal.FetchEntryByID(currenctDayId, journalEntries)
	if idx != -1 {
		fmt.Printf("There is already an entry for %s. Do you want to override it? (y/N): ", dateStr)
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput != "y" {
			fmt.Println("No changes made...")
			return nil
		}
		journalEntries[idx] = *journal.NewJournalEntry()
		fmt.Printf("Data for %s overwrote.", dateStr)
		return nil
	}
	newEntry := journal.NewJournalEntry()
	journalEntries = append(journalEntries, *newEntry)
	fmt.Printf("Added new Journal Entry for %s\n", dateStr)
	return journal.SaveEntries(journalEntries)
}

func getUserInput() (string, error) {

	var userInput string
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		return "", err
	}
	return strings.ToLower(userInput), nil
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
