package journal

import (
	"fmt"
	"time"
)

type JournalEntry struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Notes     []string
}

func NewJournalEntry() *JournalEntry {
	now := time.Now()
	id := now.Format("20060102")
	return &JournalEntry{ID: id, StartTime: time.Now()}
}

func (j *JournalEntry) String() string {
	start := j.StartTime.Format("15:04:05")
	end := j.EndTime.Format("15:04:05")
	totalTime := j.EndTime.Sub(j.StartTime).String()
	if j.EndTime.IsZero() {
		end = "Not yet closed"
		totalTime = "N/A"
	}
	timeStr := fmt.Sprintf("Start: %s | End: %s | Time: %s", start, end, totalTime)
	notes := ""
	for _, note := range j.Notes {
		notes += fmt.Sprintf("- %s\n", note)
	}
	headerStr := fmt.Sprintf("Date: %s", j.StartTime.Format("2006-01-02"))
	return fmt.Sprintf("%s\n%s\n\n%s", headerStr, timeStr, notes)
}

func (j *JournalEntry) AddNote(note string) error {
	if note == "" {
		return fmt.Errorf("Cannot add empty note")
	}
	j.Notes = append(j.Notes, note)
	return nil
}

func (j *JournalEntry) EndDay() {
	j.EndTime = time.Now()
}

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
