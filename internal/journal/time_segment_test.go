package journal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTimeSegmentDuration(t *testing.T) {
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 11, 30, 0, 0, time.UTC)

	t.Run("completed segment returns correct duration", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			EndTime:   endTime,
		}
		expected := 2*time.Hour + 30*time.Minute
		if segment.Duration() != expected {
			t.Errorf("Expected duration %v, got %v", expected, segment.Duration())
		}
	})

	t.Run("ongoing segment returns zero duration", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			// EndTime is zero (ongoing)
		}
		if segment.Duration() != 0 {
			t.Errorf("Expected duration 0 for ongoing segment, got %v", segment.Duration())
		}
	})
}

func TestTimeSegmentIsActive(t *testing.T) {
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 11, 30, 0, 0, time.UTC)

	t.Run("completed segment is not active", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			EndTime:   endTime,
		}
		if segment.IsActive() {
			t.Error("Expected completed segment to not be active")
		}
	})

	t.Run("ongoing segment is active", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			// EndTime is zero (ongoing)
		}
		if !segment.IsActive() {
			t.Error("Expected ongoing segment to be active")
		}
	})
}

func TestTimeSegmentGetClient(t *testing.T) {
	t.Run("returns client when set", func(t *testing.T) {
		segment := TimeSegment{
			Client: "client1",
		}
		if segment.GetClient() != "client1" {
			t.Errorf("Expected client1, got %s", segment.GetClient())
		}
	})

	t.Run("returns general when client is empty", func(t *testing.T) {
		segment := TimeSegment{
			Client: "",
		}
		if segment.GetClient() != "general" {
			t.Errorf("Expected general, got %s", segment.GetClient())
		}
	})

	t.Run("returns general when client is not set", func(t *testing.T) {
		segment := TimeSegment{}
		if segment.GetClient() != "general" {
			t.Errorf("Expected general, got %s", segment.GetClient())
		}
	})
}

func TestTimeSegmentString(t *testing.T) {
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 11, 30, 0, 0, time.UTC)

	t.Run("completed segment string representation", func(t *testing.T) {
		segment := TimeSegment{
			ID:          "1",
			StartTime:   startTime,
			EndTime:     endTime,
			Client:      "client1",
			Project:     "aws-migration",
			Task:        "ec2-rightsizing",
			Description: "Cost analysis",
		}
		expected := "[completed] client1/aws-migration/ec2-rightsizing (2h30m0s) - Cost analysis"
		result := segment.String()
		if result != expected {
			t.Errorf("Expected: %s, got: %s", expected, result)
		}
	})

	t.Run("active segment string representation", func(t *testing.T) {
		segment := TimeSegment{
			ID:          "1",
			StartTime:   startTime,
			Project:     "aws-migration",
			Task:        "ec2-rightsizing",
			Description: "Cost analysis",
		}
		expected := "[active] general/aws-migration/ec2-rightsizing (ongoing) - Cost analysis"
		result := segment.String()
		if result != expected {
			t.Errorf("Expected: %s, got: %s", expected, result)
		}
	})
}

func TestJournalEntryAddTimeSegment(t *testing.T) {
	entry := NewJournalEntry()
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)

	t.Run("adds valid time segment", func(t *testing.T) {
		segment := TimeSegment{
			StartTime:   startTime,
			Project:     "aws-migration",
			Task:        "ec2-rightsizing",
			Description: "Cost analysis",
		}

		err := entry.AddTimeSegment(segment)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(entry.TimeSegments) != 1 {
			t.Errorf("Expected 1 time segment, got %d", len(entry.TimeSegments))
		}

		// Check that ID was auto-generated
		if entry.TimeSegments[0].ID != "1" {
			t.Errorf("Expected ID '1', got '%s'", entry.TimeSegments[0].ID)
		}
	})

	t.Run("rejects invalid time segment", func(t *testing.T) {
		invalidSegment := TimeSegment{
			StartTime: startTime,
			// Missing required Project and Task
		}

		err := entry.AddTimeSegment(invalidSegment)
		if err == nil {
			t.Error("Expected error for invalid segment")
		}
	})

	t.Run("preserves custom ID if provided", func(t *testing.T) {
		entry := NewJournalEntry() // Fresh entry
		segment := TimeSegment{
			ID:        "custom-id",
			StartTime: startTime,
			Project:   "test-project",
			Task:      "test-task",
		}

		err := entry.AddTimeSegment(segment)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if entry.TimeSegments[0].ID != "custom-id" {
			t.Errorf("Expected ID 'custom-id', got '%s'", entry.TimeSegments[0].ID)
		}
	})
}

func TestJournalEntryGetActiveTimeSegments(t *testing.T) {
	entry := NewJournalEntry()
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 11, 0, 0, 0, time.UTC)

	// Add completed segment
	completedSegment := TimeSegment{
		ID:        "1",
		StartTime: startTime,
		EndTime:   endTime,
		Project:   "project1",
		Task:      "task1",
	}
	entry.AddTimeSegment(completedSegment)

	// Add active segment
	activeSegment := TimeSegment{
		ID:        "2",
		StartTime: startTime.Add(2 * time.Hour),
		Project:   "project2",
		Task:      "task2",
	}
	entry.AddTimeSegment(activeSegment)

	activeSegments := entry.GetActiveTimeSegments()

	if len(activeSegments) != 1 {
		t.Errorf("Expected 1 active segment, got %d", len(activeSegments))
	}

	if activeSegments[0].ID != "2" {
		t.Errorf("Expected active segment ID '2', got '%s'", activeSegments[0].ID)
	}
}

func TestJournalEntryStopTimeSegment(t *testing.T) {
	entry := NewJournalEntry()
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)

	// Add active segment
	activeSegment := TimeSegment{
		ID:        "1",
		StartTime: startTime,
		Project:   "project1",
		Task:      "task1",
	}
	entry.AddTimeSegment(activeSegment)

	t.Run("stops active segment successfully", func(t *testing.T) {
		err := entry.StopTimeSegment("1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if entry.TimeSegments[0].IsActive() {
			t.Error("Expected segment to be stopped")
		}

		if entry.TimeSegments[0].EndTime.IsZero() {
			t.Error("Expected EndTime to be set")
		}
	})

	t.Run("fails to stop already stopped segment", func(t *testing.T) {
		err := entry.StopTimeSegment("1")
		if err == nil {
			t.Error("Expected error when stopping already stopped segment")
		}
	})

	t.Run("fails to stop non-existent segment", func(t *testing.T) {
		err := entry.StopTimeSegment("999")
		if err == nil {
			t.Error("Expected error when stopping non-existent segment")
		}
	})
}

func TestJournalEntryGetTimeSegmentsByProject(t *testing.T) {
	entry := NewJournalEntry()
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)

	// Add segments for different projects
	segment1 := TimeSegment{
		ID:        "1",
		StartTime: startTime,
		Project:   "aws-migration",
		Task:      "task1",
	}
	segment2 := TimeSegment{
		ID:        "2",
		StartTime: startTime,
		Project:   "build-optimization",
		Task:      "task2",
	}
	segment3 := TimeSegment{
		ID:        "3",
		StartTime: startTime,
		Project:   "aws-migration",
		Task:      "task3",
	}

	entry.AddTimeSegment(segment1)
	entry.AddTimeSegment(segment2)
	entry.AddTimeSegment(segment3)

	awsSegments := entry.GetTimeSegmentsByProject("aws-migration")
	if len(awsSegments) != 2 {
		t.Errorf("Expected 2 aws-migration segments, got %d", len(awsSegments))
	}

	buildSegments := entry.GetTimeSegmentsByProject("build-optimization")
	if len(buildSegments) != 1 {
		t.Errorf("Expected 1 build-optimization segment, got %d", len(buildSegments))
	}

	nonExistentSegments := entry.GetTimeSegmentsByProject("non-existent")
	if len(nonExistentSegments) != 0 {
		t.Errorf("Expected 0 non-existent segments, got %d", len(nonExistentSegments))
	}
}

func TestJournalEntryGetTimeSegmentsByClient(t *testing.T) {
	entry := NewJournalEntry()
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)

	// Add segments for different clients
	segment1 := TimeSegment{
		ID:        "1",
		StartTime: startTime,
		Client:    "client1",
		Project:   "project1",
		Task:      "task1",
	}
	segment2 := TimeSegment{
		ID:        "2",
		StartTime: startTime,
		Client:    "client2",
		Project:   "project2",
		Task:      "task2",
	}
	segment3 := TimeSegment{
		ID:        "3",
		StartTime: startTime,
		// No client set (should default to "general")
		Project: "project3",
		Task:    "task3",
	}

	entry.AddTimeSegment(segment1)
	entry.AddTimeSegment(segment2)
	entry.AddTimeSegment(segment3)

	client1Segments := entry.GetTimeSegmentsByClient("client1")
	if len(client1Segments) != 1 {
		t.Errorf("Expected 1 client1 segment, got %d", len(client1Segments))
	}

	generalSegments := entry.GetTimeSegmentsByClient("general")
	if len(generalSegments) != 1 {
		t.Errorf("Expected 1 general segment, got %d", len(generalSegments))
	}
}

func TestValidateTimeSegment(t *testing.T) {
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 9, 10, 0, 0, time.UTC) // 10 minutes later

	t.Run("valid segment passes validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			EndTime:   endTime,
			Project:   "aws-migration",
			Task:      "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err != nil {
			t.Errorf("Expected no error for valid segment, got %v", err)
		}
	})

	t.Run("segment without project fails validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			Task:      "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err == nil {
			t.Error("Expected error for segment without project")
		}
	})

	t.Run("segment without task fails validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			Project:   "aws-migration",
		}

		err := ValidateTimeSegment(segment)
		if err == nil {
			t.Error("Expected error for segment without task")
		}
	})

	t.Run("segment without start time fails validation", func(t *testing.T) {
		segment := TimeSegment{
			Project: "aws-migration",
			Task:    "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err == nil {
			t.Error("Expected error for segment without start time")
		}
	})

	t.Run("segment with end time before start time fails validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: endTime,
			EndTime:   startTime, // End before start
			Project:   "aws-migration",
			Task:      "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err == nil {
			t.Error("Expected error for segment with end time before start time")
		}
	})

	t.Run("segment with duration less than 5 minutes fails validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			EndTime:   startTime.Add(2 * time.Minute), // Only 2 minutes
			Project:   "aws-migration",
			Task:      "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err == nil {
			t.Error("Expected error for segment with duration less than 5 minutes")
		}
	})

	t.Run("ongoing segment (no end time) passes validation", func(t *testing.T) {
		segment := TimeSegment{
			StartTime: startTime,
			// No EndTime set (ongoing)
			Project: "aws-migration",
			Task:    "ec2-rightsizing",
		}

		err := ValidateTimeSegment(segment)
		if err != nil {
			t.Errorf("Expected no error for ongoing segment, got %v", err)
		}
	})
}

func TestTimeSegmentJSONSerialization(t *testing.T) {
	startTime := time.Date(2025, 10, 27, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 10, 27, 11, 30, 0, 0, time.UTC)

	t.Run("time segment serializes correctly", func(t *testing.T) {
		segment := TimeSegment{
			ID:          "1",
			StartTime:   startTime,
			EndTime:     endTime,
			Client:      "client1",
			Project:     "aws-migration",
			Task:        "ec2-rightsizing",
			Description: "Cost analysis",
		}

		data, err := json.Marshal(segment)
		if err != nil {
			t.Errorf("Expected no error marshaling segment, got %v", err)
		}

		var decoded TimeSegment
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Expected no error unmarshaling segment, got %v", err)
		}

		if !cmp.Equal(segment, decoded) {
			t.Errorf("Segments don't match after JSON roundtrip:\n%s", cmp.Diff(segment, decoded))
		}
	})

	t.Run("journal entry with time segments serializes correctly", func(t *testing.T) {
		entry := &JournalEntry{
			ID:        "20251027",
			StartTime: startTime,
			EndTime:   endTime,
			TimeSegments: []TimeSegment{
				{
					ID:          "1",
					StartTime:   startTime,
					EndTime:     startTime.Add(2 * time.Hour),
					Client:      "client1",
					Project:     "aws-migration",
					Task:        "ec2-rightsizing",
					Description: "Cost analysis",
				},
				{
					ID:          "2",
					StartTime:   startTime.Add(2 * time.Hour),
					EndTime:     endTime,
					Project:     "build-optimization",
					Task:        "pipeline-analysis",
					Description: "Performance review",
				},
			},
		}

		data, err := json.Marshal(entry)
		if err != nil {
			t.Errorf("Expected no error marshaling entry, got %v", err)
		}

		var decoded JournalEntry
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Expected no error unmarshaling entry, got %v", err)
		}

		if len(decoded.TimeSegments) != 2 {
			t.Errorf("Expected 2 time segments, got %d", len(decoded.TimeSegments))
		}

		if decoded.TimeSegments[0].Project != "aws-migration" {
			t.Errorf("Expected first project to be aws-migration, got %s", decoded.TimeSegments[0].Project)
		}

		if decoded.TimeSegments[1].GetClient() != "general" {
			t.Errorf("Expected second client to be general, got %s", decoded.TimeSegments[1].GetClient())
		}
	})
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("existing JSON without time_segments loads properly", func(t *testing.T) {
		// This represents existing JSON structure without time_segments field
		oldJSON := `{
			"id": "20251027",
			"start_time": "2025-10-27T09:00:00Z",
			"end_time": "2025-10-27T17:00:00Z",
			"notes": [],
			"breaks": []
		}`

		var entry JournalEntry
		err := json.Unmarshal([]byte(oldJSON), &entry)
		if err != nil {
			t.Errorf("Expected no error loading old format, got %v", err)
		}

		// TimeSegments should be nil (default zero value) when field is missing from JSON
		// This is the expected Go behavior and is perfectly fine
		if len(entry.TimeSegments) != 0 {
			t.Errorf("Expected 0 time segments, got %d", len(entry.TimeSegments))
		}
	})

	t.Run("new JSON with time_segments loads properly", func(t *testing.T) {
		newJSON := `{
			"id": "20251027",
			"start_time": "2025-10-27T09:00:00Z",
			"end_time": "2025-10-27T17:00:00Z",
			"notes": [],
			"breaks": [],
			"time_segments": [
				{
					"id": "1",
					"start_time": "2025-10-27T09:00:00Z",
					"end_time": "2025-10-27T11:00:00Z",
					"client": "client1",
					"project": "aws-migration",
					"task": "ec2-rightsizing",
					"description": "Cost analysis"
				}
			]
		}`

		var entry JournalEntry
		err := json.Unmarshal([]byte(newJSON), &entry)
		if err != nil {
			t.Errorf("Expected no error loading new format, got %v", err)
		}

		if len(entry.TimeSegments) != 1 {
			t.Errorf("Expected 1 time segment, got %d", len(entry.TimeSegments))
		}

		segment := entry.TimeSegments[0]
		if segment.Project != "aws-migration" {
			t.Errorf("Expected project aws-migration, got %s", segment.Project)
		}

		if segment.GetClient() != "client1" {
			t.Errorf("Expected client1, got %s", segment.GetClient())
		}
	})
}
