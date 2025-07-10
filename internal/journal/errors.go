package journal

import (
	"errors"
	"fmt"
)

// Error types for better error categorization and handling
var (
	// ErrEmptyNote is returned when attempting to add an empty note
	ErrEmptyNote = errors.New("cannot add empty note")
	
	// ErrNoEntries is returned when no entries are found in operations requiring entries
	ErrNoEntries = errors.New("no entries found")
	
	// ErrInvalidEntry is returned when an entry has invalid data
	ErrInvalidEntry = errors.New("invalid entry")
	
	// ErrEntryNotFound is returned when a specific entry cannot be found
	ErrEntryNotFound = errors.New("entry not found")
	
	// ErrInvalidTimeFormat is returned when time parsing fails
	ErrInvalidTimeFormat = errors.New("invalid time format")
	
	// ErrInvalidBreak is returned when break operations fail
	ErrInvalidBreak = errors.New("invalid break operation")
	
	// ErrJournalIO is returned when journal file operations fail
	ErrJournalIO = errors.New("journal file operation failed")
	
	// ErrValidation is returned when validation fails
	ErrValidation = errors.New("validation failed")
)

// JournalError represents a structured error with context
type JournalError struct {
	Type    error
	Message string
	Context map[string]interface{}
	Err     error
}

func (e *JournalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Type.Error(), e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type.Error(), e.Message)
}

func (e *JournalError) Unwrap() error {
	return e.Err
}

// Is allows error comparison
func (e *JournalError) Is(target error) bool {
	return errors.Is(e.Type, target)
}

// WithContext adds context to the error
func (e *JournalError) WithContext(key string, value interface{}) *JournalError {
	e.Context[key] = value
	return e
}

// Helper functions for common error scenarios

// EmptyNoteError creates an error for empty note operations
func EmptyNoteError() error {
	return &JournalError{
		Type:    ErrEmptyNote,
		Message: "note content cannot be empty",
		Context: make(map[string]interface{}),
	}
}

// NoEntriesError creates an error for operations requiring entries
func NoEntriesError(operation string) error {
	return &JournalError{
		Type:    ErrNoEntries,
		Message: fmt.Sprintf("no entries found for %s", operation),
		Context: make(map[string]interface{}),
	}
}

// EntryNotFoundError creates an error for missing entries
func EntryNotFoundError(id string) error {
	return &JournalError{
		Type:    ErrEntryNotFound,
		Message: fmt.Sprintf("entry with id %s not found", id),
		Context: make(map[string]interface{}),
	}
}

// InvalidEntryError creates an error for invalid entries
func InvalidEntryError(id string, reason string) error {
	return &JournalError{
		Type:    ErrInvalidEntry,
		Message: fmt.Sprintf("entry %s is invalid: %s", id, reason),
		Context: make(map[string]interface{}),
	}
}

// TimeFormatError creates an error for time parsing issues
func TimeFormatError(input string, err error) error {
	return &JournalError{
		Type:    ErrInvalidTimeFormat,
		Message: fmt.Sprintf("failed to parse time '%s'", input),
		Context: make(map[string]interface{}),
		Err:     err,
	}
}

// BreakError creates an error for break operations
func BreakError(reason string) error {
	return &JournalError{
		Type:    ErrInvalidBreak,
		Message: reason,
		Context: make(map[string]interface{}),
	}
}

// JournalIOError creates an error for file operations
func JournalIOError(operation string, err error) error {
	return &JournalError{
		Type:    ErrJournalIO,
		Message: fmt.Sprintf("failed to %s journal", operation),
		Context: make(map[string]interface{}),
		Err:     err,
	}
}

// ValidationError creates an error for validation failures
func ValidationError(field string, reason string) error {
	return &JournalError{
		Type:    ErrValidation,
		Message: fmt.Sprintf("validation failed for %s: %s", field, reason),
		Context: make(map[string]interface{}),
	}
}