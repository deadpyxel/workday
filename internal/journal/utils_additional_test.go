package journal

import (
	"testing"
	"time"
)

func TestFetchEntriesByMonthDate(t *testing.T) {
	// Create test entries for different months
	jan1 := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	jan15 := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	feb1 := time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC)
	mar1 := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	
	entries := []JournalEntry{
		{ID: "20240101", StartTime: jan1},
		{ID: "20240115", StartTime: jan15},
		{ID: "20240201", StartTime: feb1},
		{ID: "20240301", StartTime: mar1},
	}

	tests := []struct {
		name           string
		entries        []JournalEntry
		filterDate     time.Time
		expectedCount  int
		expectError    bool
	}{
		{
			name:           "January entries",
			entries:        entries,
			filterDate:     jan1,
			expectedCount:  2,
			expectError:    false,
		},
		{
			name:           "February entries",
			entries:        entries,
			filterDate:     feb1,
			expectedCount:  1,
			expectError:    false,
		},
		{
			name:           "March entries",
			entries:        entries,
			filterDate:     mar1,
			expectedCount:  1,
			expectError:    false,
		},
		{
			name:           "No entries for April",
			entries:        entries,
			filterDate:     time.Date(2024, 4, 1, 9, 0, 0, 0, time.UTC),
			expectedCount:  0,
			expectError:    true,
		},
		{
			name:           "Empty entries slice",
			entries:        []JournalEntry{},
			filterDate:     jan1,
			expectedCount:  0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FetchEntriesByMonthDate(tt.entries, tt.filterDate)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got nil")
				}
				if len(result) != 0 {
					t.Errorf("Expected empty slice when error occurs, got %d entries", len(result))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("Expected %d entries, got %d", tt.expectedCount, len(result))
				}
				
				// Verify all returned entries are from the correct month
				expectedYear, expectedMonth := tt.filterDate.Year(), tt.filterDate.Month()
				for _, entry := range result {
					entryYear, entryMonth, _ := entry.StartTime.Date()
					if entryYear != expectedYear || entryMonth != expectedMonth {
						t.Errorf("Entry %s is not from the expected month %d/%d", entry.ID, expectedYear, expectedMonth)
					}
				}
			}
		})
	}
}

func TestBreakDuration(t *testing.T) {
	now := time.Now()
	later := now.Add(30 * time.Minute)
	
	tests := []struct {
		name     string
		br       Break
		expected time.Duration
	}{
		{
			name:     "valid break with 30 minutes",
			br:       Break{StartTime: now, EndTime: later, Reason: "coffee"},
			expected: 30 * time.Minute,
		},
		{
			name:     "break without end time",
			br:       Break{StartTime: now, EndTime: time.Time{}, Reason: "ongoing"},
			expected: 0,
		},
		{
			name:     "break with 1 hour duration",
			br:       Break{StartTime: now, EndTime: now.Add(time.Hour), Reason: "lunch"},
			expected: time.Hour,
		},
		{
			name:     "break with 15 minutes duration",
			br:       Break{StartTime: now, EndTime: now.Add(15 * time.Minute), Reason: "quick break"},
			expected: 15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.br.Duration()
			if result != tt.expected {
				t.Errorf("Duration() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestJournalEntryTotalWorkTime(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		entry    JournalEntry
		expected time.Duration
	}{
		{
			name: "entry without end time",
			entry: JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   time.Time{},
			},
			expected: 0,
		},
		{
			name: "entry with 8 hours no breaks",
			entry: JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   now.Add(8 * time.Hour),
				Breaks:    []Break{},
			},
			expected: 8 * time.Hour,
		},
		{
			name: "entry with 8 hours and 1 hour break",
			entry: JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   now.Add(9 * time.Hour),
				Breaks: []Break{
					{StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour), Reason: "lunch"},
				},
			},
			expected: 8 * time.Hour,
		},
		{
			name: "entry with multiple breaks",
			entry: JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   now.Add(9 * time.Hour),
				Breaks: []Break{
					{StartTime: now.Add(2 * time.Hour), EndTime: now.Add(2*time.Hour + 15*time.Minute), Reason: "coffee"},
					{StartTime: now.Add(4 * time.Hour), EndTime: now.Add(5 * time.Hour), Reason: "lunch"},
				},
			},
			expected: 7*time.Hour + 45*time.Minute,
		},
		{
			name: "entry with ongoing break",
			entry: JournalEntry{
				ID:        "20240101",
				StartTime: now,
				EndTime:   now.Add(8 * time.Hour),
				Breaks: []Break{
					{StartTime: now.Add(4 * time.Hour), EndTime: time.Time{}, Reason: "ongoing"},
				},
			},
			expected: 8 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.TotalWorkTime()
			if result != tt.expected {
				t.Errorf("TotalWorkTime() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestJournalEntryEndDay(t *testing.T) {
	entry := JournalEntry{
		ID:        "20240101",
		StartTime: time.Now(),
		EndTime:   time.Time{},
	}
	
	// EndTime should be zero initially
	if !entry.EndTime.IsZero() {
		t.Error("Expected EndTime to be zero initially")
	}
	
	// Call EndDay
	entry.EndDay()
	
	// EndTime should now be set
	if entry.EndTime.IsZero() {
		t.Error("Expected EndTime to be set after EndDay()")
	}
	
	// EndTime should be after StartTime
	if !entry.EndTime.After(entry.StartTime) {
		t.Error("Expected EndTime to be after StartTime")
	}
}

func TestNewJournalEntry(t *testing.T) {
	entry := NewJournalEntry()
	
	if entry == nil {
		t.Fatal("Expected non-nil entry")
	}
	
	if entry.ID == "" {
		t.Error("Expected ID to be set")
	}
	
	if entry.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
	
	if !entry.EndTime.IsZero() {
		t.Error("Expected EndTime to be zero for new entry")
	}
	
	// ID should be in YYYYMMDD format
	expectedID := time.Now().Format("20060102")
	if entry.ID != expectedID {
		t.Errorf("Expected ID to be %s, got %s", expectedID, entry.ID)
	}
}