package journal

import (
	"fmt"
	"time"
)

// FetchEntryByID searches for a JournalEntry in the provided slice of entries by its ID.
// It returns a pointer to the found JournalEntry and its index in the slice
// If the entry is not found, it returns nil and -1.
func FetchEntryByID(id string, entries []JournalEntry) (*JournalEntry, int) {
	for i, entry := range entries {
		if entry.ID == id {
			return &entry, i
		}
	}
	return nil, -1
}

func CurrentWeekEntries(journalEntries []JournalEntry) ([]JournalEntry, error) {
	if journalEntries == nil {
		return []JournalEntry{}, fmt.Errorf("No entries were passed")
	}
	var currentWeekEntries []JournalEntry

	// Get the current year and ISO week number.
	now := time.Now()
	currentYear, currentWeek := now.ISOWeek()

	for _, entry := range journalEntries {
		// get the year and ISO week number of the entry.
		entryYear, entryWeek := entry.StartTime.ISOWeek()

		// If the entry's year and week match the current year and week, add the entry to the slice.
		if entryYear == currentYear && entryWeek == currentWeek {
			currentWeekEntries = append(currentWeekEntries, entry)
		}
	}
	if currentWeekEntries == nil {
		return currentWeekEntries, fmt.Errorf("No entries found for the current week")
	}

	return currentWeekEntries, nil
}
