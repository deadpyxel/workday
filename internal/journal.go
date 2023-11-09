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
	for i, note := range j.Notes {
		notes += fmt.Sprintf("- %s", note)
		// Only append newline character if the note is not the last one
		if i < len(j.Notes)-1 {
			notes += "\n"
		}
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
