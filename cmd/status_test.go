package cmd

import (
	"testing"
	"time"

	"github.com/deadpyxel/workday/internal/journal"
)

func TestCalculateExpectedEndTime(t *testing.T) {
	tests := []struct {
		name             string
		entry            *journal.JournalEntry
		minWorkTime      time.Duration
		lunchTime        time.Duration
		now              time.Time
		expectedEndTime  time.Time
		expectedWorkTime time.Duration
	}{
		{
			name: "No breaks taken, lunch break needed",
			entry: &journal.JournalEntry{
				ID:        "20240101",
				StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				Breaks:    []journal.Break{},
			},
			minWorkTime:      8*time.Hour + 20*time.Minute,
			lunchTime:        1 * time.Hour,
			now:              time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			expectedEndTime:  time.Date(2024, 1, 1, 18, 20, 0, 0, time.UTC), // 9:00 + 8h20m + 1h lunch
			expectedWorkTime: 2 * time.Hour,                                  // 2 hours worked so far
		},
		{
			name: "Lunch break completed",
			entry: &journal.JournalEntry{
				ID:        "20240101",
				StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				Breaks: []journal.Break{
					{
						StartTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
						Reason:    "lunch",
					},
				},
			},
			minWorkTime:      8*time.Hour + 20*time.Minute,
			lunchTime:        1 * time.Hour,
			now:              time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			expectedEndTime:  time.Date(2024, 1, 1, 18, 20, 0, 0, time.UTC), // 9:00 + 8h20m + 1h break
			expectedWorkTime: 4 * time.Hour,                                  // 5 hours total - 1 hour break
		},
		{
			name: "Multiple breaks with lunch",
			entry: &journal.JournalEntry{
				ID:        "20240101",
				StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				Breaks: []journal.Break{
					{
						StartTime: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
						EndTime:   time.Date(2024, 1, 1, 10, 45, 0, 0, time.UTC),
						Reason:    "coffee",
					},
					{
						StartTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
						Reason:    "lunch",
					},
				},
			},
			minWorkTime:      8*time.Hour + 20*time.Minute,
			lunchTime:        1 * time.Hour,
			now:              time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			expectedEndTime:  time.Date(2024, 1, 1, 18, 35, 0, 0, time.UTC), // 9:00 + 8h20m + 1h15m breaks
			expectedWorkTime: 4*time.Hour + 45*time.Minute,                  // 6 hours total - 1h15m breaks
		},
		{
			name: "Short breaks only, lunch needed",
			entry: &journal.JournalEntry{
				ID:        "20240101",
				StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				Breaks: []journal.Break{
					{
						StartTime: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
						EndTime:   time.Date(2024, 1, 1, 10, 45, 0, 0, time.UTC),
						Reason:    "coffee",
					},
					{
						StartTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
						EndTime:   time.Date(2024, 1, 1, 14, 15, 0, 0, time.UTC),
						Reason:    "coffee",
					},
				},
			},
			minWorkTime:      8*time.Hour + 20*time.Minute,
			lunchTime:        1 * time.Hour,
			now:              time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
			expectedEndTime:  time.Date(2024, 1, 1, 18, 50, 0, 0, time.UTC), // 9:00 + 8h20m + 30m breaks + 1h lunch
			expectedWorkTime: 6*time.Hour + 30*time.Minute,                  // 7 hours total - 30m breaks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedEndTime, timeRemaining, currentWorkTime := calculateExpectedEndTime(tt.entry, tt.minWorkTime, tt.lunchTime, tt.now)
			
			if !expectedEndTime.Equal(tt.expectedEndTime) {
				t.Errorf("calculateExpectedEndTime() expectedEndTime = %v, want %v", expectedEndTime, tt.expectedEndTime)
			}
			
			if currentWorkTime != tt.expectedWorkTime {
				t.Errorf("calculateExpectedEndTime() currentWorkTime = %v, want %v", currentWorkTime, tt.expectedWorkTime)
			}
			
			expectedTimeRemaining := tt.expectedEndTime.Sub(tt.now)
			if expectedTimeRemaining < 0 {
				expectedTimeRemaining = 0
			}
			
			if timeRemaining != expectedTimeRemaining {
				t.Errorf("calculateExpectedEndTime() timeRemaining = %v, want %v", timeRemaining, expectedTimeRemaining)
			}
		})
	}
}

func TestCalculateExpectedEndTimeWithOngoingBreak(t *testing.T) {
	entry := &journal.JournalEntry{
		ID:        "20240101",
		StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		Breaks: []journal.Break{
			{
				StartTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
				Reason:    "lunch",
			},
			{
				StartTime: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
				// EndTime is zero (ongoing break)
				Reason: "coffee",
			},
		},
	}

	minWorkTime := 8*time.Hour + 20*time.Minute
	lunchTime := 1 * time.Hour
	now := time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC) // Currently on break

	expectedEndTime, timeRemaining, currentWorkTime := calculateExpectedEndTime(entry, minWorkTime, lunchTime, now)

	// Should only count completed breaks (1 hour lunch)
	expectedEnd := time.Date(2024, 1, 1, 18, 20, 0, 0, time.UTC) // 9:00 + 8h20m + 1h lunch
	if !expectedEndTime.Equal(expectedEnd) {
		t.Errorf("calculateExpectedEndTime() expectedEndTime = %v, want %v", expectedEndTime, expectedEnd)
	}

	// Current work time should be 6 hours (up to start of ongoing break) - 1 hour lunch
	expectedWorkTime := 5 * time.Hour
	if currentWorkTime != expectedWorkTime {
		t.Errorf("calculateExpectedEndTime() currentWorkTime = %v, want %v", currentWorkTime, expectedWorkTime)
	}

	// Time remaining should be positive
	if timeRemaining <= 0 {
		t.Errorf("calculateExpectedEndTime() timeRemaining = %v, should be positive", timeRemaining)
	}
}