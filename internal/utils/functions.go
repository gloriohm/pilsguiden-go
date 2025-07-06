package utils

import (
	"errors"
	"go-router/internal/stores"
	"go-router/models"
	"sort"
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
