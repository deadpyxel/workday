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

// FetchEntriesByWeekDate filters a slice of JournalEntry objects and returns a new slice
// containing only the entries from the current week. The function uses the ISO week date
// system, where weeks start on a Monday and the first week of the year is the one that
// includes at least four days of the new year.
//
// The function takes a slice of JournalEntry objects and iterates over each entry. It
// checks the start time of each entry and compares it with the current week. If the entry
// belongs to the current week, it is added to the new slice.
//
// If no entries are passed, or no entries belong to the current week, the function returns
// an error along with an empty slice.
//
// Example:
//
//	entries := []JournalEntry{...}
//	currentWeekEntries, err := FetchEntriesByWeekDate(entries)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(currentWeekEntries) // Prints the entries for the current week
func FetchEntriesByWeekDate(journalEntries []JournalEntry, currentDate time.Time) ([]JournalEntry, error) {
	if len(journalEntries) == 0 {
		return nil, fmt.Errorf("No entries were passed")
	}
	var currentWeekEntries []JournalEntry

	// Get the current year and ISO week number.
	currentYear, currentWeek := currentDate.ISOWeek()

	for _, entry := range journalEntries {
		// get the year and ISO week number of the entry.
		entryYear, entryWeek := entry.StartTime.ISOWeek()

		// If the entry's year and week match the current year and week, add the entry to the slice.
		if entryYear == currentYear && entryWeek == currentWeek {
			currentWeekEntries = append(currentWeekEntries, entry)
		}
	}
	if len(currentWeekEntries) == 0 {
		return nil, fmt.Errorf("No entries found for the current week")
	}

	return currentWeekEntries, nil
}
