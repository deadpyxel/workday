package journal

import (
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
// containing only the entries from the given date's week. The function uses the ISO week date
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
		return nil, NoEntriesError("weekly report")
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
		return nil, NoEntriesError("current week")
	}

	return currentWeekEntries, nil
}

// FetchEntriesByMonthDate filters a slice of JournalEntry objects and returns a new slice
// containing only the entries from the given date's month. The function uses the Year() and Month()
// functions from time standard library
//
// The function takes a slice of JournalEntry objects and iterates over each entry. It
// checks the start time of each entry and compares it with the current month. If the entry
// belongs to the current month, it is added to the new slice.
//
// If no entries are passed, or no entries belong to the current month, the function returns
// an error along with an empty slice.
//
// Example:
//
//	entries := []JournalEntry{...}
//	currentMonthEntries, err := FetchEntriesByMonthDate(entries)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(currentMonthEntries) // Prints the entries for the current week
func FetchEntriesByMonthDate(journalEntries []JournalEntry, filterDate time.Time) ([]JournalEntry, error) {
	if len(journalEntries) == 0 {
		return nil, NoEntriesError("monthly report")
	}
	var currentMonthEntries []JournalEntry

	// Get the current year and month.
	currentYear, currentMonth := filterDate.Year(), filterDate.Month()

	for _, entry := range journalEntries {
		// get the year and month of the entry.
		entryYear, entryMonth, _ := entry.StartTime.Date()

		// If the entry's year and month match the current year and month, add the entry to the slice.
		if entryYear == currentYear && entryMonth == currentMonth {
			currentMonthEntries = append(currentMonthEntries, entry)
		}
	}
	if len(currentMonthEntries) == 0 {
		return nil, NoEntriesError("current month")
	}

	return currentMonthEntries, nil
}

// CalculateTotalTime calculates the total time duration for a slice of JournalEntry.
// It iterates over each entry in the slice and checks if the end time of the entry is after the start time.
// If the end time is not after the start time, it returns an error indicating that the entry is invalid.
// Otherwise, it adds the difference between the end time and the start time to the total duration.
// After iterating over all entries, it returns the total duration and nil.
// If there are no entries in the slice, it returns a duration of 0 and nil.
func CalculateTotalTime(entries []JournalEntry) (time.Duration, error) {
	var d time.Duration
	for _, entry := range entries {
		if !entry.EndTime.After(entry.StartTime) {
			return 0, InvalidEntryError(entry.ID, "end time is before start time")
		}
		d += entry.EndTime.Sub(entry.StartTime)
	}
	return d, nil
}
