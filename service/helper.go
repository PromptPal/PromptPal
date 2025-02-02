package service

import "strings"

// replacePlaceholders replaces variables in the prompt template with their values
func replacePlaceholders(prompt string, variables map[string]string) string {
	result := prompt
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}
