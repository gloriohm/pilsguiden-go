package utils

import (
	"errors"
	"go-router/internal/stores"
	"go-router/models"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func SetNavParams(params map[string]string) (models.UrlNav, error) {
	var nav models.UrlNav
	fylke, ok := params["fylke"]
	if ok {
		loc, err := stores.AppStore.GetLocationBySlug(fylke, "fylke")
		if err != nil {
			return nav, errors.New("fylke not found")
		} else {
			nav.Fylke = loc
		}
	}
	kommune, ok := params["kommune"]
	if ok {
		loc, err := stores.AppStore.GetLocationBySlug(kommune, "kommune")
		if err != nil {
			return nav, errors.New("kommune not found")
		} else {
			nav.Kommune = loc
		}
	}
	sted, ok := params["sted"]
	if ok {
		loc, err := stores.AppStore.GetLocationBySlug(sted, "sted")
		if err != nil {
			return nav, errors.New("sted not found")
		} else {
			nav.Sted = loc
		}
	}
	return nav, nil
}

func ExtractSortedUniqueKommuner(bars []models.Bar) []models.BaseLocation {
	seen := make(map[int]models.BaseLocation)
	for _, bar := range bars {
		if _, exists := seen[bar.Kommune]; !exists {
			seen[bar.Kommune] = models.BaseLocation{
				Slug: bar.KommuneSlug,
				Name: bar.KommuneName,
			}
		}
	}

	// Flatten map to slice
	var urls []models.BaseLocation
	for _, url := range seen {
		urls = append(urls, url)
	}

	// Sort by Name
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].Name < urls[j].Name
	})

	return urls
}

func ExtractSortedUniqueSteder(bars []models.Bar) []models.BaseLocation {
	seen := make(map[int]models.BaseLocation)
	for _, bar := range bars {
		if bar.Sted == nil {
			continue // skip bars without a sted
		}
		id := *bar.Sted
		if _, exists := seen[id]; !exists {
			seen[id] = models.BaseLocation{
				Slug: *bar.StedSlug,
				Name: *bar.StedName,
			}
		}
	}

	// Flatten to slice
	var urls []models.BaseLocation
	for _, url := range seen {
		urls = append(urls, url)
	}

	// Sort by Name
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].Name < urls[j].Name
	})

	return urls
}

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
