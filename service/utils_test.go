package service

import (
	"testing"
)

func TestReplacePlaceholdersLegacy(t *testing.T) {
	tests := []struct {
		name         string
		sentence     string
		replacements map[string]string
		expected     string
	}{
		{
			name:         "Basic replacement",
			sentence:     "Hello {{name}}!",
			replacements: map[string]string{"name": "John"},
			expected:     "Hello John!",
		},
		{
			name:         "Multiple replacements",
			sentence:     "{{greeting}} {{name}}, how are you?",
			replacements: map[string]string{"greeting": "Hi", "name": "Alice"},
			expected:     "Hi Alice, how are you?",
		},
		{
			name:         "No matching replacement",
			sentence:     "Hello {{unknown}}!",
			replacements: map[string]string{"name": "John"},
			expected:     "Hello {{unknown}}!",
		},
		{
			name:         "With whitespace in placeholder",
			sentence:     "Hello {{  name  }}!",
			replacements: map[string]string{"name": "John"},
			expected:     "Hello John!",
		},
		{
			name:         "Empty replacements map",
			sentence:     "Hello {{name}}!",
			replacements: map[string]string{},
			expected:     "Hello {{name}}!",
		},
		{
			name:         "No placeholders",
			sentence:     "Hello world!",
			replacements: map[string]string{"name": "John"},
			expected:     "Hello world!",
		},
		{
			name:         "Multiple occurrences of same placeholder",
			sentence:     "{{name}} is {{name}}",
			replacements: map[string]string{"name": "John"},
			expected:     "John is John",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacePlaceholdersLegacy(tt.sentence, tt.replacements)
			if result != tt.expected {
				t.Errorf("replacePlaceholdersLegacy() = %v, want %v", result, tt.expected)
			}
		})
	}
}
