package unittest

import "unicode"

// NormalizeFileName converts all non-digit, non-letter runes in input string to
// dash ('-'). Coalesces multiple dashes into one.
func NormalizeFileName(s string) string {
	var result []rune
	for _, r := range []rune(s) {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			result = append(result, r)
		} else {
			l := len(result)
			if l > 0 && result[l-1] != '-' {
				result = append(result, rune('-'))
			}
		}
	}
	return string(result)
}
