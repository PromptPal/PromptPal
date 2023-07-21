package service

import (
	"regexp"
)

func replacePlaceholders(sentence string, replacements map[string]string) string {
	re := regexp.MustCompile(`{{\s*([a-zA-Z][a-zA-Z0-9]*)\s*}}`)
	result := re.ReplaceAllStringFunc(sentence, func(match string) string {
		placeholder := re.FindStringSubmatch(match)[1]
		if value, ok := replacements[placeholder]; ok {
			return value
		}
		return match
	})
	return result
}
