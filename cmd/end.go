package cmd

import (
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
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
	dateStr := now.Format("2006-01-02")

	// Load entries
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	// Get the index of the entry on the slice, -1 if not found
	entry, idx := journal.FetchEntryByID(currentDayId, entries)

	if idx == -1 {
		return fmt.Errorf("No entry found for the current day")
	}

	if !entry.EndTime.IsZero() {
		fmt.Printf("There is already an EndTime for %s. Do you want to override it? (y/N): ", dateStr)
		userInput, err := getUserInput()
		if err != nil {
			return err
		}
		if userInput != "y" {
			fmt.Println("No changes made...")
			return nil
		}
		entry.EndDay()
		entries[idx] = *entry
		fmt.Printf("Data for %s overwrote. Saving...", dateStr)
	} else {
		entries[idx].EndDay()
	}

	err = validateEntry(entry)
	if err != nil {
		fmt.Printf("\nFound issues with entry: %v", err)
		validationNote := journal.Note{Contents: fmt.Sprintf("Validation Error: %s", err)}
		entry.AddNote(validationNote)
	}

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return fmt.Errorf("Failed to save journal entries: %v\n", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(endCmd)
}

func validateEntry(entry *journal.JournalEntry) error {
	minWorkTime, err := time.ParseDuration(viper.GetString("minWorkTime"))
	if err != nil {
		return fmt.Errorf("invalid minimum work time format in config: %v", err)
	}
	lunchTime, err := time.ParseDuration(viper.GetString("lunchTime"))
	if err != nil {
		return fmt.Errorf("invalid lunch time format in config: %v", err)
	}
	maxWorkTime, err := time.ParseDuration(viper.GetString("maxWorkTime"))
	if err != nil {
		return fmt.Errorf("invalid maximum work time format in config: %v", err)
	}

	// Check if total work time (accounting for breaks) is less than minimum
	if entry.TotalWorkTime() < minWorkTime {
		return fmt.Errorf("total work time (%s) is less than the minimum required (%s)", entry.TotalWorkTime().String(), minWorkTime.String())
	}
	// Check if total work time (accounting for breaks) above allowed maximum
	if entry.TotalWorkTime() > maxWorkTime {
		return fmt.Errorf("total work time (%s) exceeds the maximum allowed (%s) by %s", entry.TotalWorkTime().String(), maxWorkTime.String(), entry.TotalWorkTime()-maxWorkTime)
	}

	// Check if there's at least one break of `lunchtime` duration
	for _, br := range entry.Breaks {
		if br.Duration() >= lunchTime {
			return nil
		}
	}
	// If not, return an error
	return fmt.Errorf("did not find any breaks during the day that have at least (%s) for lunchtime break.", lunchTime)
}
