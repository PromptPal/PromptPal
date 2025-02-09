package service

import (
	"testing"
)

func TestReplacePlaceholders(t *testing.T) {
	prompt := `
	You are a knowledgeable and highly educated assistant with extensive reading experience. I will provide you with a key-value pair that contains information about the book I am reading and the passage I want to understand. Your task is to help me explain it clearly and concisely.\n\nYou have to translate all your answers to {{ lang }}\n\n------\n\ntitle: {{ bookTitle }}\nauthor: {{ author }}\npublish date: {{ pbDate }}\nURL: {{ url }}\nISBN: {{ isbn }}\nbook summary: {{ summary }}\npassage: {{ passage }}
	`

	result := replacePlaceholdersLegacy(prompt, map[string]string{
		"lang":      "English",
		"bookTitle": "The Great Gatsby",
		"author":    "F. Scott Fitzgerald",
		"pbDate":    "1925",
		"url":       "https://example.com/gatsby",
		"isbn":      "978-0-306-40615-7",
		"summary":   "The Great Gatsby is a 1925 novel written by F. Scott Fitzgerald.",
		"passage":   "In my younger and more vulnerable years my father gave me some advice that I've been turning over in my mind ever since.",
	})

	expected := `
	You are a knowledgeable and highly educated assistant with extensive reading experience. I will provide you with a key-value pair that contains information about the book I am reading and the passage I want to understand. Your task is to help me explain it clearly and concisely.\n\nYou have to translate all your answers to English\n\n------\n\ntitle: The Great Gatsby\nauthor: F. Scott Fitzgerald\npublish date: 1925\nURL: https://example.com/gatsby\nISBN: 978-0-306-40615-7\nbook summary: The Great Gatsby is a 1925 novel written by F. Scott Fitzgerald.\npassage: In my younger and more vulnerable years my father gave me some advice that I've been turning over in my mind ever since.
	`

	if result != expected {
		t.Errorf("Expected result to be empty, but got: %s", result)
	}
}

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
