package utils

import (
	"go-router/models"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func ToURL(input string) string {
	// Lowercase
	s := strings.ToLower(input)

	// Replace "&" with "og"
	s = strings.ReplaceAll(s, "&", "og")

	// Replace æ, ø, å with ae, o, a
	s = strings.ReplaceAll(s, "æ", "ae")
	s = strings.ReplaceAll(s, "ø", "o")
	s = strings.ReplaceAll(s, "å", "a")

	// Replace all whitespace with hyphens
	spaceToHyphen := regexp.MustCompile(`\s+`)
	s = spaceToHyphen.ReplaceAllString(s, "-")

	// Normalize and strip diacritics
	t := norm.NFD.String(s)
	sanitized := make([]rune, 0, len(t))
	for _, r := range t {
		// Skip diacritical marks
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		// Allow a–z, 0–9, hyphen
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sanitized = append(sanitized, r)
		}
	}

	// Collapse multiple hyphens into one
	hyphenCollapse := regexp.MustCompile(`-+`)
	result := hyphenCollapse.ReplaceAllString(string(sanitized), "-")

	// Trim leading/trailing hyphens
	result = strings.Trim(result, "-")

	return result
}

func CheckValidLocationLevel(level string) bool {
	switch level {
	case "fylke":
		return true
	case "kommune":
		return true
	case "sted":
		return true
	default:
		return false
	}
}

func ToBase(in []models.Location) []models.BaseLocation {
	out := make([]models.BaseLocation, len(in))
	for i := range in {
		out[i] = in[i].BaseLocation
	}
	return out
}
