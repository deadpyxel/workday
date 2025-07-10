package journal

import (
	"strings"
	"time"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid bool
	Error   error
}

// ValidateNote validates a note's content and tags
func ValidateNote(note Note) ValidationResult {
	if strings.TrimSpace(note.Contents) == "" {
		return ValidationResult{
			IsValid: false,
			Error:   EmptyNoteError(),
		}
	}

	// Clean up tags - remove empty strings
	validTags := make([]string, 0, len(note.Tags))
	for _, tag := range note.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			validTags = append(validTags, trimmed)
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateEntry validates a journal entry
func ValidateEntry(entry *JournalEntry) ValidationResult {
	if entry == nil {
		return ValidationResult{
			IsValid: false,
			Error:   ValidationError("entry", "entry is nil"),
		}
	}

	if entry.ID == "" {
		return ValidationResult{
			IsValid: false,
			Error:   ValidationError("id", "entry ID cannot be empty"),
		}
	}

	if entry.StartTime.IsZero() {
		return ValidationResult{
			IsValid: false,
			Error:   ValidationError("start_time", "start time cannot be zero"),
		}
	}

	// If end time is set, validate it's after start time
	if !entry.EndTime.IsZero() && !entry.EndTime.After(entry.StartTime) {
		return ValidationResult{
			IsValid: false,
			Error:   InvalidEntryError(entry.ID, "end time must be after start time"),
		}
	}

	// Validate all breaks
	for i, br := range entry.Breaks {
		if result := ValidateBreak(br); !result.IsValid {
			return ValidationResult{
				IsValid: false,
				Error:   ValidationError("breaks", "break "+string(rune(i))+" is invalid: "+result.Error.Error()),
			}
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// ValidateBreak validates a break entry
func ValidateBreak(br Break) ValidationResult {
	if br.StartTime.IsZero() {
		return ValidationResult{
			IsValid: false,
			Error:   ValidationError("break_start_time", "break start time cannot be zero"),
		}
	}

	if !br.EndTime.IsZero() && !br.EndTime.After(br.StartTime) {
		return ValidationResult{
			IsValid: false,
			Error:   BreakError("break end time must be after start time"),
		}
	}

	if strings.TrimSpace(br.Reason) == "" {
		return ValidationResult{
			IsValid: false,
			Error:   BreakError("break reason cannot be empty"),
		}
	}

	return ValidationResult{IsValid: true, Error: nil}
}

// FindCurrentDayEntry finds the entry for the current day
func FindCurrentDayEntry(entries []JournalEntry) (*JournalEntry, int, error) {
	if len(entries) == 0 {
		return nil, -1, NoEntriesError("current day lookup")
	}

	currentDayId := time.Now().Format("20060102")
	entry, idx := FetchEntryByID(currentDayId, entries)
	if idx == -1 {
		return nil, -1, EntryNotFoundError(currentDayId)
	}

	return entry, idx, nil
}

// ValidateTimeFormat validates and parses a time string in HH:MM format
func ValidateTimeFormat(timeStr string) (time.Time, error) {
	if strings.TrimSpace(timeStr) == "" {
		return time.Time{}, ValidationError("time", "time string cannot be empty")
	}

	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, TimeFormatError(timeStr, err)
	}

	return parsedTime, nil
}

// ValidateConfigDuration validates and parses a duration string from config
func ValidateConfigDuration(durationStr string, fieldName string) (time.Duration, error) {
	if strings.TrimSpace(durationStr) == "" {
		return 0, ValidationError(fieldName, "duration cannot be empty")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, ValidationError(fieldName, "invalid duration format: "+err.Error())
	}

	return duration, nil
}