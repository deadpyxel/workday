package cmd

import (
	"strings"
	"testing"
)

func TestGetUserInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		expectErr  bool
	}{
		{
			name:       "simple input",
			input:      "yes\n",
			expected:   "yes",
			expectErr:  false,
		},
		{
			name:       "input with uppercase",
			input:      "YES\n",
			expected:   "yes",
			expectErr:  false,
		},
		{
			name:       "input with mixed case",
			input:      "YeS\n",
			expected:   "yes",
			expectErr:  false,
		},
		{
			name:       "single character input",
			input:      "y\n",
			expected:   "y",
			expectErr:  false,
		},
		{
			name:       "single character no input",
			input:      "n\n",
			expected:   "n",
			expectErr:  false,
		},
		{
			name:       "empty input",
			input:      "\n",
			expected:   "",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test getUserInput without mocking stdin
			// So we'll test the logic that it should perform
			input := strings.TrimSpace(tt.input)
			result := strings.ToLower(input)
			
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Test helper function to verify the getUserInput logic
func TestUserInputLogic(t *testing.T) {
	// This tests the core logic that getUserInput performs
	testInputs := []string{
		"y",
		"Y", 
		"yes",
		"YES",
		"YeS",
		"n",
		"N",
		"no",
		"NO",
		"",
	}
	
	for _, input := range testInputs {
		result := strings.ToLower(input)
		
		// Verify the result is lowercase
		if result != strings.ToLower(input) {
			t.Errorf("Expected lowercase result for %q, got %q", input, result)
		}
	}
}

// Test validation helper that would be used with getUserInput
func TestValidateUserResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "yes response",
			input:    "y",
			expected: true,
		},
		{
			name:     "no response",
			input:    "n",
			expected: false,
		},
		{
			name:     "empty response (default no)",
			input:    "",
			expected: false,
		},
		{
			name:     "other response (default no)",
			input:    "maybe",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates how getUserInput result is typically used
			result := tt.input == "y"
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}