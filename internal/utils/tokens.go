package utils

import (
	"fmt"
	"strings"
)

// EstimateTokens provides a rough estimate of the number of tokens in the given text.
// This is based on the heuristic that one token is approximately four characters.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	cleanText := strings.ReplaceAll(text, "\n", " ")
	return len(cleanText) / 4
}

// FormatTokenCount formats a token count with K/M suffixes for readability
func FormatTokenCount(count int) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(count)/1000000)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.1fk", float64(count)/1000)
	}
	return fmt.Sprintf("%d", count)
}
