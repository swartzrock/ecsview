package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Renders an int64 to a string
func I64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Renders a string to lower title case (all lower case except for initial chars in each word)
func LowerTitle(s string) string {
	return strings.Title(strings.ToLower(s))
}

// Replaces all regex matches in a string
func ReplaceAllRegex(pattern string, src string, replace string) string {
	return regexp.MustCompile(pattern).ReplaceAllString(src, replace)
}

// Removes all regex matches in a string
func RemoveAllRegex(pattern string, src string) string {
	return ReplaceAllRegex(pattern, src, "")
}

// Returns the right x chars from a string
func TakeRight(s string, max int) string {
	result := s
	if len(result) > max {
		if max > 1 {
			max -= 1
		}
		start := len(result) - max - 1
		result = "…" + s[start:len(result)]
	}
	return result
}

// Returns the left x chars from a string
func TakeLeft(s string, max int) string {
	result := s
	if len(result) > max {
		if max > 1 {
			max -= 1
		}
		result = s[0:max] + "…"
	}
	return result
}

// Builds a one-line meter using the amount and total values limited to the given width
func BuildAsciiMeterCurrentTotal(portion int64, total int64, width int) string {
	const fullChar = "█"
	const emptyChar = "▒"

	full := 0
	if total > 0 {
		ratio := float64(portion) / float64(total)
		ratio = math.Max(0, math.Min(1.0, ratio))
		full = int(math.Round(ratio * float64(width)))
	}

	return strings.Join([]string{
		strings.Repeat(fullChar, full),
		strings.Repeat(emptyChar, width-full),
	}, "")
}
