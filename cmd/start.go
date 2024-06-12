package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
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
// If the last entry on the journal was not closed properly (misisng EndTime) it will offer to update that first
// If there is already an entry for the current day, it asks the user if they want to override it.
// If the user agrees, it overwrites the existing entry with a new one.
// If the user does not agree, it does nothing and returns nil.
// If there is no entry for the current day, it creates a new JournalEntry with the current date and time as the start time,
// appends the new entry to the journal entries, and saves the updated journal entries back to the file.
// It then prints a message indicating that a new JournalEntry has been added for the current day.
func startWorkDay(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	var lastEntry *journal.JournalEntry
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].ID[:8] != currentDayId {
			lastEntry = &entries[i]
			break
		}
	}
	// If the last entry has no EndTime set (considering it is not the entry for the current day)
	if lastEntry != nil && lastEntry.EndTime.IsZero() {
		fmt.Println("Warning: Last entry has no EndTime set.")
		fmt.Printf("Do you want to set the EndTime for the last entry? (y/N): ")
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput == "y" {
			fmt.Printf("Please type the EndTime in HH:MM format: ")
			endTimeStr, err := getUserInput()
			if err != nil {
				return err
			}
			endTime, err := time.Parse("15:04", endTimeStr)
			if err != nil {
				return err
			}
			finalEndTime := time.Date(
				lastEntry.StartTime.Year(),
				lastEntry.StartTime.Month(),
				lastEntry.StartTime.Day(),
				endTime.Hour(),
				endTime.Minute(),
				0,
				0,
				lastEntry.StartTime.Location())
			lastEntry.EndTime = finalEndTime

			err = journal.SaveEntries(entries, journalPath)
			if err != nil {
				return err
			}

			fmt.Println("Endtime set for the last entry.")
		}
	}

	dateStr := now.Format("2006-01-02")
	_, idx := journal.FetchEntryByID(currentDayId, entries)
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
		entries[idx] = *journal.NewJournalEntry()
		fmt.Printf("Data for %s overwrote. Saving...", dateStr)
		return journal.SaveEntries(entries, journalPath)
	}
	newEntry := journal.NewJournalEntry()
	entries = append(entries, *newEntry)
	fmt.Printf("Added new Journal Entry for %s\n", dateStr)
	return journal.SaveEntries(entries, journalPath)
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
