package cmd

import (
	"fmt"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var breakReason string

var breakCmd = &cobra.Command{
	Use:   "break",
	Short: "Manages work break entries",
	Long:  "The break command allows you to start and stop tracking work breaks.",
}

var breakStartCmd = &cobra.Command{
	Use:   "start [reason]",
	Short: "starts a new work break",
	Long:  "Starts a new work break, recording the start time and reason",
	Args:  cobra.MinimumNArgs(1), // Reason is mandatory
	RunE:  startBreak,
}

func startBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	entry, idx := journal.FetchEntryByID(currentDayId, entries)
	if idx == -1 {
		return fmt.Errorf("No entry found for the current day. Start your workday first.")
	}

	if len(args) > 0 {
		breakReason = args[0]
	}

	newBreak := journal.Break{
		StartTime: now,
		Reason:    breakReason,
	}

	entry.Breaks = append(entry.Breaks, newBreak)
	entries[idx] = *entry // Update the entry in the slice

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	fmt.Printf("Break started at %s\n", now.Format("15:04:05"))
	return nil
}

var breakStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops a current work break",
	Long:  "Stops the current work break, recording the end time.",
	RunE:  stopBreak,
}

func stopBreak(cmd *cobra.Command, args []string) error {
	journalPath := viper.GetString("journalPath")
	entries, err := journal.LoadEntries(journalPath)
	if err != nil {
		return err
	}

	now := time.Now()
	currentDayId := now.Format("20060102")
	entry, idx := journal.FetchEntryByID(currentDayId, entries)
	if idx == -1 {
		return fmt.Errorf("No entry found for the current day.")
	}

	if len(entry.Breaks) == 0 {
		return fmt.Errorf("No break started for the current day.")
	}

	lastBreak := &entry.Breaks[len(entry.Breaks)-1] // Get the last break

	if !lastBreak.EndTime.IsZero() {
		return fmt.Errorf("Last break was already stopped.")
	}
	lastBreak.EndTime = now
	entries[idx] = *entry

	err = journal.SaveEntries(entries, journalPath)
	if err != nil {
		return err
	}

	fmt.Printf("Break stopped at %s\n", now.Format("15:04:05"))
	return nil
}

func init() {
	rootCmd.AddCommand(breakCmd)
	breakCmd.AddCommand(breakStartCmd)
	breakCmd.AddCommand(breakStopCmd)

	// Add flag to the `start` subcommand
	breakStartCmd.Flags().StringVarP(&breakReason, "reason", "r", "", "Reason for the break")
}
