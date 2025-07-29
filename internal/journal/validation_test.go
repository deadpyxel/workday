package journal

import (
	"errors"
	"testing"
	"time"
)

func TestValidateNote(t *testing.T) {
	tests := []struct {
		name     string
		note     Note
		expected bool
	}{
		{
			name:     "valid note with content",
			note:     Note{Contents: "Valid note content"},
			expected: true,
		},
		{
			name:     "valid note with content and tags",
			note:     Note{Contents: "Valid note content", Tags: []string{"tag1", "tag2"}},
			expected: true,
		},
		{
			name:     "invalid note with empty content",
			note:     Note{Contents: ""},
			expected: false,
		},
		{
			name:     "invalid note with whitespace only content",
			note:     Note{Contents: "   "},
			expected: false,
		},
		{
			name:     "valid note with empty tags",
			note:     Note{Contents: "Valid content", Tags: []string{}},
			expected: true,
		},
		{
			name:     "valid note with mixed empty and valid tags",
			note:     Note{Contents: "Valid content", Tags: []string{"", "valid-tag", "  "}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateNote(tt.note)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateNote() = %v, expected %v", result.IsValid, tt.expected)
			}
			if !tt.expected && result.Error == nil {
				t.Error("Expected error for invalid note, but got nil")
			}
		})
	}
}

func TestValidateEntry(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		entry    *JournalEntry
		expected bool
	}{
		{
			name:     "nil entry",
			entry:    nil,
			expected: false,
		},
		{
			name: "valid entry with all fields",
			entry: &JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   later,
				Notes:     []Note{{Contents: "Test note"}},
				Breaks:    []Break{{StartTime: now.Add(30 * time.Minute), EndTime: now.Add(45 * time.Minute), Reason: "coffee"}},
			},
			expected: true,
		},
		{
			name: "valid entry without end time",
			entry: &JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   time.Time{},
			},
			expected: true,
		},
		{
			name: "invalid entry with empty ID",
			entry: &JournalEntry{
				ID:        "",
				StartTime: now,
				EndTime:   later,
			},
			expected: false,
		},
		{
			name: "invalid entry with zero start time",
			entry: &JournalEntry{
				ID:        "20240101",
				StartTime: time.Time{},
				EndTime:   later,
			},
			expected: false,
		},
		{
			name: "invalid entry with end time before start time",
			entry: &JournalEntry{
				ID:        "20240101",
				StartTime: later,
				EndTime:   now,
			},
			expected: false,
		},
		{
			name: "invalid entry with invalid break",
			entry: &JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   later,
				Breaks:    []Break{{StartTime: time.Time{}, EndTime: now, Reason: "invalid"}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEntry(tt.entry)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateEntry() = %v, expected %v", result.IsValid, tt.expected)
			}
			if !tt.expected && result.Error == nil {
				t.Error("Expected error for invalid entry, but got nil")
			}
		})
	}
}

func TestValidateBreak(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		br       Break
		expected bool
	}{
		{
			name:     "valid break with all fields",
			br:       Break{StartTime: now, EndTime: later, Reason: "lunch"},
			expected: true,
		},
		{
			name:     "valid break without end time",
			br:       Break{StartTime: now, EndTime: time.Time{}, Reason: "coffee"},
			expected: true,
		},
		{
			name:     "invalid break with zero start time",
			br:       Break{StartTime: time.Time{}, EndTime: later, Reason: "lunch"},
			expected: false,
		},
		{
			name:     "invalid break with end time before start time",
			br:       Break{StartTime: later, EndTime: now, Reason: "lunch"},
			expected: false,
		},
		{
			name:     "invalid break with empty reason",
			br:       Break{StartTime: now, EndTime: later, Reason: ""},
			expected: false,
		},
		{
			name:     "invalid break with whitespace only reason",
			br:       Break{StartTime: now, EndTime: later, Reason: "   "},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateBreak(tt.br)
			if result.IsValid != tt.expected {
				t.Errorf("ValidateBreak() = %v, expected %v", result.IsValid, tt.expected)
			}
			if !tt.expected && result.Error == nil {
				t.Error("Expected error for invalid break, but got nil")
			}
		})
	}
}

func TestFindCurrentDayEntry(t *testing.T) {
	now := time.Now()
	currentDayId := now.Format("20060102")
	
	tests := []struct {
		name        string
		entries     []JournalEntry
		expectEntry bool
		expectError bool
	}{
		{
			name:        "empty entries",
			entries:     []JournalEntry{},
			expectEntry: false,
			expectError: true,
		},
		{
			name: "current day entry exists",
			entries: []JournalEntry{
				{ID: currentDayId, StartTime: now},
			},
			expectEntry: true,
			expectError: false,
		},
		{
			name: "current day entry does not exist",
			entries: []JournalEntry{
				{ID: "20240101", StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)},
			},
			expectEntry: false,
			expectError: true,
		},
		{
			name: "multiple entries with current day",
			entries: []JournalEntry{
				{ID: "20240101", StartTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)},
				{ID: currentDayId, StartTime: now},
				{ID: "20240102", StartTime: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)},
			},
			expectEntry: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, idx, err := FindCurrentDayEntry(tt.entries)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
				if entry != nil {
					t.Error("Expected nil entry when error occurs")
				}
				if idx != -1 {
					t.Error("Expected index -1 when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !tt.expectEntry && entry != nil {
					t.Error("Expected nil entry, but got non-nil")
				}
				if tt.expectEntry && entry == nil {
					t.Error("Expected non-nil entry, but got nil")
				}
				if tt.expectEntry && entry != nil && entry.ID != currentDayId {
					t.Errorf("Expected entry ID %s, got %s", currentDayId, entry.ID)
				}
			}
		})
	}
}

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		expectErr bool
	}{
		{
			name:      "valid time format",
			timeStr:   "09:30",
			expectErr: false,
		},
		{
			name:      "valid time format with leading zero",
			timeStr:   "08:00",
			expectErr: false,
		},
		{
			name:      "valid time format 24h",
			timeStr:   "23:59",
			expectErr: false,
		},
		{
			name:      "invalid time format with seconds",
			timeStr:   "09:30:45",
			expectErr: true,
		},
		{
			name:      "invalid time format single digit",
			timeStr:   "9:30",
			expectErr: false,
		},
		{
			name:      "empty time string",
			timeStr:   "",
			expectErr: true,
		},
		{
			name:      "whitespace only time string",
			timeStr:   "   ",
			expectErr: true,
		},
		{
			name:      "invalid time format with AM/PM",
			timeStr:   "09:30 AM",
			expectErr: true,
		},
		{
			name:      "invalid time with wrong separator",
			timeStr:   "09.30",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateTimeFormat(tt.timeStr)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
				if !result.IsZero() {
					t.Error("Expected zero time when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsZero() {
					t.Error("Expected valid time, but got zero time")
				}
			}
		})
	}
}

func TestValidateConfigDuration(t *testing.T) {
	tests := []struct {
		name        string
		durationStr string
		fieldName   string
		expectErr   bool
	}{
		{
			name:        "valid duration in hours",
			durationStr: "8h",
			fieldName:   "workTime",
			expectErr:   false,
		},
		{
			name:        "valid duration in minutes",
			durationStr: "30m",
			fieldName:   "breakTime",
			expectErr:   false,
		},
		{
			name:        "valid duration in seconds",
			durationStr: "45s",
			fieldName:   "shortBreak",
			expectErr:   false,
		},
		{
			name:        "valid complex duration",
			durationStr: "2h30m",
			fieldName:   "longBreak",
			expectErr:   false,
		},
		{
			name:        "empty duration string",
			durationStr: "",
			fieldName:   "workTime",
			expectErr:   true,
		},
		{
			name:        "whitespace only duration",
			durationStr: "   ",
			fieldName:   "workTime",
			expectErr:   true,
		},
		{
			name:        "invalid duration format",
			durationStr: "8 hours",
			fieldName:   "workTime",
			expectErr:   true,
		},
		{
			name:        "invalid duration with wrong unit",
			durationStr: "8x",
			fieldName:   "workTime",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateConfigDuration(tt.durationStr, tt.fieldName)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
				if result != 0 {
					t.Error("Expected zero duration when error occurs")
				}
				// Check if error is a ValidationError
				var validationErr *JournalError
				if errors.As(err, &validationErr) {
					if !errors.Is(validationErr, ErrValidation) {
						t.Error("Expected ValidationError type")
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == 0 {
					t.Error("Expected non-zero duration, but got zero")
				}
			}
		})
	}
}