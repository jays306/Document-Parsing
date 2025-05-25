package parsers

import (
	"testing"
	"strings"
)

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No markdown",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "With markdown code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "With markdown code block without language",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "With extra whitespace",
			input:    "  \n  {\"key\": \"value\"}  \n  ",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("cleanJSONResponse(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJobDetailsPrompt(t *testing.T) {
	prompt := jobDetailsPrompt()
	
	// Check that the prompt contains the expected fields
	expectedFields := []string{
		"title",
		"salary",
		"location",
		"experience",
		"employment-type",
	}
	
	for _, field := range expectedFields {
		if !strings.Contains(prompt, field) {
			t.Errorf("jobDetailsPrompt() does not contain field %q", field)
		}
	}
}

func TestForm941Prompt(t *testing.T) {
	prompt := form941Prompt()
	
	// Check that the prompt contains the expected fields
	expectedFields := []string{
		"EIN",
		"Name",
		"Trade name",
		"Address",
		"Box 1",
		"Box 4",
		"Box 5e",
	}
	
	for _, field := range expectedFields {
		if !strings.Contains(prompt, field) {
			t.Errorf("form941Prompt() does not contain field %q", field)
		}
	}
}