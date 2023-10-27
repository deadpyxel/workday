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
	totalTime := j.EndTime.Sub(j.StartTime)
	timeStr := fmt.Sprintf("Start: %s | End: %s | Time: %s", start, end, totalTime.String())
	notes := ""
	for _, note := range j.Notes {
		notes += fmt.Sprintf("- %s\n", note)
	}
	return fmt.Sprintf("\n%s\n\n%s", timeStr, notes)
}

func (j *JournalEntry) AddNote(note string) {
	j.Notes = append(j.Notes, note)
}

func (j *JournalEntry) EndDay() {
	j.EndTime = time.Now()
}
