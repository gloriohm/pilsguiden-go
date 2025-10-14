package main

import (
	"context"
	"go-router/internal/bars"
	"go-router/internal/stores"
	"go-router/models"
	"go-router/templates"
	"net/http"
	"sort"
)

func generateErrorPage(w http.ResponseWriter, ctx context.Context, title string, consented bool) {
	templates.Layout(title, consented, templates.ErrorPage()).Render(ctx, w)
}

func setNavParams(fylke, kommune, sted string) (models.UrlNav, models.CurrentLvl, error) {
	var nav models.UrlNav
	var current models.CurrentLvl
	if fylke != "" {
		fylkePath := "/" + fylke
		loc, err := stores.AppStore.GetLocationBySlug(fylkePath, "fylke")
		if err != nil {
			return nav, current, err
		} else {
			nav.Fylke = loc
			current.Name = "fylke"
			current.ID = loc.ID
		}
	}

	if kommune != "" {
		kommunePath := "/" + fylke + "/" + kommune
		loc, err := stores.AppStore.GetLocationBySlug(kommunePath, "kommune")
		if err != nil {
			return nav, current, err
		} else {
			nav.Kommune = loc
			current.Name = "kommune"
			current.ID = loc.ID
		}
	}

	if sted != "" {
		stedPath := "/" + fylke + "/" + kommune + "/" + sted
		loc, err := stores.AppStore.GetLocationBySlug(stedPath, "sted")
		if err != nil {
			return nav, current, err
		} else {
			nav.Sted = loc
			current.Name = "sted"
			current.ID = loc.ID
		}
	}
	return nav, current, nil
}

func extractNextLocs(bars []models.BarView, lvl string) []bars.BaseLocation {
	seen := make(map[int]bars.BaseLocation)

	switch lvl {
	case "fylke":
		for _, bar := range bars {
			if _, exists := seen[bar.Kommune]; !exists {
				seen[bar.Kommune] = bars.BaseLocation{
					Slug: bar.KommuneSlug,
					Name: bar.KommuneName,
				}
			}
		}

	case "kommune", "sted":
		for _, bar := range bars {
			if bar.Sted == nil || bar.StedSlug == nil || bar.StedName == nil {
				continue
			}
			id := *bar.Sted
			if _, exists := seen[id]; !exists {
				seen[id] = bars.BaseLocation{
					Slug: *bar.StedSlug,
					Name: *bar.StedName,
				}
			}
		}

	default:
		return []bars.BaseLocation{}
	}

	locs := make([]bars.BaseLocation, 0, len(seen))
	for _, loc := range seen {
		locs = append(locs, loc)
	}
	sort.Slice(locs, func(i, j int) bool { return locs[i].Name < locs[j].Name })

	return locs
}
