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

// ParseContent extracts hashtags from the note content and updates both
// the Contents and Tags fields. This allows users to write notes with
// inline tags like "Meeting done #progress #team" and have them automatically
// parsed into separate fields.
func (n *Note) ParseContent() {
	cleanContent, tags := ParseNoteTags(n.Contents)
	n.Contents = cleanContent

	// Merge with existing tags, avoiding duplicates
	existingTags := make(map[string]bool)
	for _, tag := range n.Tags {
		existingTags[tag] = true
	}

	for _, tag := range tags {
		if !existingTags[tag] {
			n.Tags = append(n.Tags, tag)
		}
	}
}

type Break struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

func (b *Break) Duration() time.Duration {
	var total time.Duration
	if b.EndTime.IsZero() {
		return total
	}
	total = b.EndTime.Sub(b.StartTime)
	return total
}

type TimeSegment struct {
	ID          string    `json:"id"`                    // Unique ID within the day
	StartTime   time.Time `json:"start_time"`            // When tracking started
	EndTime     time.Time `json:"end_time,omitempty"`    // When tracking stopped (omitempty for ongoing)
	Client      string    `json:"client,omitempty"`      // Optional client name, defaults to "general"
	Project     string    `json:"project"`               // Required project name
	Task        string    `json:"task"`                  // Required task name
	Description string    `json:"description,omitempty"` // Optional description
}

// Duration calculates the duration of a time segment
func (ts *TimeSegment) Duration() time.Duration {
	if ts.EndTime.IsZero() {
		return 0 // Ongoing segment has no duration yet
	}
	return ts.EndTime.Sub(ts.StartTime)
}

// IsActive returns true if the time segment is currently active (no end time)
func (ts *TimeSegment) IsActive() bool {
	return ts.EndTime.IsZero()
}

// GetClient returns the client name, defaulting to "general" if empty
func (ts *TimeSegment) GetClient() string {
	if ts.Client == "" {
		return "general"
	}
	return ts.Client
}

// String provides a human-readable representation
func (ts *TimeSegment) String() string {
	status := "completed"
	duration := ts.Duration().String()
	if ts.IsActive() {
		status = "active"
		duration = "ongoing"
	}

	client := ts.GetClient()
	return fmt.Sprintf("[%s] %s/%s/%s (%s) - %s",
		status, client, ts.Project, ts.Task, duration, ts.Description)
}

type JournalEntry struct {
	ID           string        `json:"id"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Notes        []Note        `json:"notes,omitempty"`
	Breaks       []Break       `json:"breaks,omitempty"`
	TimeSegments []TimeSegment `json:"time_segments,omitempty"` // Time tracking segments
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
		return EmptyNoteError()
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

func (j *JournalEntry) TotalWorkTime() time.Duration {
	var totalWorkTime time.Duration
	if j.EndTime.IsZero() {
		return totalWorkTime
	}
	totalWorkTime = j.EndTime.Sub(j.StartTime)

	var totalBreakTime time.Duration
	for _, br := range j.Breaks {
		if !br.EndTime.IsZero() {
			totalBreakTime += br.Duration()
		}
	}

	totalWorkTime -= totalBreakTime

	return totalWorkTime
}

// AddTimeSegment adds a new time segment to the journal entry
func (j *JournalEntry) AddTimeSegment(segment TimeSegment) error {
	// Generate ID if not provided
	if segment.ID == "" {
		segment.ID = fmt.Sprintf("%d", len(j.TimeSegments)+1)
	}

	// Validate the segment
	if err := ValidateTimeSegment(segment); err != nil {
		return err
	}

	j.TimeSegments = append(j.TimeSegments, segment)
	return nil
}

// GetActiveTimeSegments returns all currently active (ongoing) time segments
func (j *JournalEntry) GetActiveTimeSegments() []TimeSegment {
	var active []TimeSegment
	for _, segment := range j.TimeSegments {
		if segment.IsActive() {
			active = append(active, segment)
		}
	}
	return active
}

// StopTimeSegment stops a time segment by ID
func (j *JournalEntry) StopTimeSegment(segmentID string) error {
	for i := range j.TimeSegments {
		if j.TimeSegments[i].ID == segmentID {
			if !j.TimeSegments[i].IsActive() {
				return fmt.Errorf("time segment %s is already stopped", segmentID)
			}
			j.TimeSegments[i].EndTime = time.Now()
			return nil
		}
	}
	return fmt.Errorf("time segment %s not found", segmentID)
}

// GetTimeSegmentsByProject returns all segments for a specific project
func (j *JournalEntry) GetTimeSegmentsByProject(project string) []TimeSegment {
	var segments []TimeSegment
	for _, segment := range j.TimeSegments {
		if segment.Project == project {
			segments = append(segments, segment)
		}
	}
	return segments
}

// GetTimeSegmentsByClient returns all segments for a specific client
func (j *JournalEntry) GetTimeSegmentsByClient(client string) []TimeSegment {
	var segments []TimeSegment
	for _, segment := range j.TimeSegments {
		if segment.GetClient() == client {
			segments = append(segments, segment)
		}
	}
	return segments
}
