package journal

import (
	"fmt"
	"time"
)

type Note struct {
	Contents string   `json:"Contents"`       // Note contents
	Tags     []string `json:"Tags,omitempty"` // Tags for this particular note
}

func (n *Note) String() string {
	tags := ""
	if len(n.Tags) > 0 {
		tags = fmt.Sprintf(" %v", n.Tags)
	}
	return fmt.Sprintf("- %s%s", n.Contents, tags)
}

type Break struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

type JournalEntry struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Notes     []Note    `json:"notes"`
	Breaks    []Break   `json:"breaks"`
}

func NewJournalEntry() *JournalEntry {
	now := time.Now()
	id := now.Format("20060102")
	return &JournalEntry{ID: id, StartTime: time.Now()}
}

type Journal struct {
	Version int            `json:"version"` // schema version for the loaded journal
	Entries []JournalEntry `json:"entries"` // journal entries
}

const SchemaVersion = 1

func (j *JournalEntry) String() string {
	start := j.StartTime.Format("15:04:05")
	end := j.EndTime.Format("15:04:05")
	var totalBreakTime time.Duration
	if len(j.Breaks) > 0 {
		for _, br := range j.Breaks {
			if br.EndTime.IsZero() {
				continue
			}
			totalBreakTime += br.EndTime.Sub(br.StartTime)
		}
	}
	totalWorkTime := j.EndTime.Sub(j.StartTime)
	totalWorkTime -= totalBreakTime
	totalTime := totalWorkTime.String()
	if j.EndTime.IsZero() {
		end = "Ongoing"
		totalTime = "N/A"
	}
	timeStr := fmt.Sprintf("Start: %s | End: %s | Time: %s", start, end, totalTime)
	notes := ""
	for i, note := range j.Notes {
		notes += note.String()
		// Only append newline character if the note is not the last one
		if i < len(j.Notes)-1 {
			notes += "\n"
		}
	}
	headerStr := fmt.Sprintf("Date: %s", j.StartTime.Format("2006-01-02"))
	return fmt.Sprintf("%s\n%s\n\n%s", headerStr, timeStr, notes)
}

func (j *JournalEntry) AddNote(note Note) error {
	if note.Contents == "" {
		return fmt.Errorf("Cannot add empty note")
	}
	if len(note.Tags) == 1 && note.Tags[0] == "" {
		note.Tags = nil
	}
	j.Notes = append(j.Notes, note)
	return nil
}

func (j *JournalEntry) EndDay() {
	j.EndTime = time.Now()
}
