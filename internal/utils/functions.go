package utils

import (
	"errors"
	"go-router/internal/stores"
	"go-router/models"
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
