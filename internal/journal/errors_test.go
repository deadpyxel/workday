package journal

import (
	"errors"
	"testing"
)

func TestJournalError(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		err      *JournalError
		expected string
	}{
		{
			name: "error with wrapped error",
			err: &JournalError{
				Type:    ErrEmptyNote,
				Message: "test message",
				Context: make(map[string]interface{}),
				Err:     baseErr,
			},
			expected: "cannot add empty note: test message - base error",
		},
		{
			name: "error without wrapped error",
			err: &JournalError{
				Type:    ErrEmptyNote,
				Message: "test message",
				Context: make(map[string]interface{}),
				Err:     nil,
			},
			expected: "cannot add empty note: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestJournalErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")

	journalErr := &JournalError{
		Type:    ErrEmptyNote,
		Message: "test message",
		Context: make(map[string]interface{}),
		Err:     baseErr,
	}

	unwrapped := journalErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, expected %v", unwrapped, baseErr)
	}
}

func TestJournalErrorIs(t *testing.T) {
	journalErr := &JournalError{
		Type:    ErrEmptyNote,
		Message: "test message",
		Context: make(map[string]interface{}),
	}

	// Test that it correctly identifies the error type
	if !journalErr.Is(ErrEmptyNote) {
		t.Error("Expected Is() to return true for ErrEmptyNote")
	}

	// Test that it returns false for different error types
	if journalErr.Is(ErrNoEntries) {
		t.Error("Expected Is() to return false for ErrNoEntries")
	}
}

func TestJournalErrorWithContext(t *testing.T) {
	journalErr := &JournalError{
		Type:    ErrEmptyNote,
		Message: "test message",
		Context: make(map[string]interface{}),
	}

	result := journalErr.WithContext("key", "value")

	if result != journalErr {
		t.Error("WithContext() should return the same instance")
	}

	if journalErr.Context["key"] != "value" {
		t.Error("WithContext() should add the key-value pair to context")
	}
}

func TestEmptyNoteError(t *testing.T) {
	err := EmptyNoteError()

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrEmptyNote) {
		t.Error("Expected error to be ErrEmptyNote")
	}
}

func TestNoEntriesError(t *testing.T) {
	operation := "test operation"
	err := NoEntriesError(operation)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrNoEntries) {
		t.Error("Expected error to be ErrNoEntries")
	}

	expectedMessage := "no entries found for " + operation
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}
}

func TestEntryNotFoundError(t *testing.T) {
	id := "20240101"
	err := EntryNotFoundError(id)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrEntryNotFound) {
		t.Error("Expected error to be ErrEntryNotFound")
	}

	expectedMessage := "entry with id " + id + " not found"
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}
}

func TestInvalidEntryError(t *testing.T) {
	id := "20240101"
	reason := "invalid reason"
	err := InvalidEntryError(id, reason)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrInvalidEntry) {
		t.Error("Expected error to be ErrInvalidEntry")
	}

	expectedMessage := "entry " + id + " is invalid: " + reason
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}
}

func TestTimeFormatError(t *testing.T) {
	input := "invalid time"
	baseErr := errors.New("parse error")
	err := TimeFormatError(input, baseErr)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrInvalidTimeFormat) {
		t.Error("Expected error to be ErrInvalidTimeFormat")
	}

	expectedMessage := "failed to parse time '" + input + "'"
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}

	if journalErr.Err != baseErr {
		t.Error("Expected wrapped error to be preserved")
	}
}

func TestBreakError(t *testing.T) {
	reason := "break error reason"
	err := BreakError(reason)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrInvalidBreak) {
		t.Error("Expected error to be ErrInvalidBreak")
	}

	if journalErr.Message != reason {
		t.Errorf("Expected message %q, got %q", reason, journalErr.Message)
	}
}

func TestJournalIOError(t *testing.T) {
	operation := "read"
	baseErr := errors.New("io error")
	err := JournalIOError(operation, baseErr)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrJournalIO) {
		t.Error("Expected error to be ErrJournalIO")
	}

	expectedMessage := "failed to " + operation + " journal"
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}

	if journalErr.Err != baseErr {
		t.Error("Expected wrapped error to be preserved")
	}
}

func TestValidationError(t *testing.T) {
	field := "test_field"
	reason := "validation reason"
	err := ValidationError(field, reason)

	if err == nil {
		t.Fatal("Expected non-nil error")
	}

	var journalErr *JournalError
	if !errors.As(err, &journalErr) {
		t.Error("Expected error to be of type JournalError")
	}

	if !errors.Is(err, ErrValidation) {
		t.Error("Expected error to be ErrValidation")
	}

	expectedMessage := "validation failed for " + field + ": " + reason
	if journalErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, journalErr.Message)
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined
	constants := []error{
		ErrEmptyNote,
		ErrNoEntries,
		ErrInvalidEntry,
		ErrEntryNotFound,
		ErrInvalidTimeFormat,
		ErrInvalidBreak,
		ErrJournalIO,
		ErrValidation,
	}

	for i, constant := range constants {
		if constant == nil {
			t.Errorf("Error constant %d is nil", i)
		}
		if constant.Error() == "" {
			t.Errorf("Error constant %d has empty message", i)
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that errors.Is works correctly with wrapped errors
	baseErr := errors.New("base error")
	journalErr := &JournalError{
		Type:    ErrEmptyNote,
		Message: "test message",
		Context: make(map[string]interface{}),
		Err:     baseErr,
	}

	// Test that we can identify the journal error type
	if !errors.Is(journalErr, ErrEmptyNote) {
		t.Error("Expected errors.Is to work with JournalError")
	}

	// Test that we can unwrap to the base error
	if !errors.Is(journalErr, baseErr) {
		t.Error("Expected errors.Is to work with wrapped error")
	}
}

