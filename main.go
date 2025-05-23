package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go-router/database"
	"go-router/models"
	"go-router/templates"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/form"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout("Home", templates.Home()).Render(r.Context(), w)
	})
	r.Get("/about", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout("About", templates.About()).Render(r.Context(), w)
	})
	r.Get("/bar/{slug}", handleBar)

	r.Route("/admin", func(r chi.Router) {
		r.Get("/create-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Create Bar", templates.BarManualForm()).Render(r.Context(), w)
		})
		r.Post("/fetch-osm", handleCreateBar)
	})

	http.ListenAndServe(":3000", r)
}

func handleBar(w http.ResponseWriter, r *http.Request) {
	barParam := chi.URLParam(r, "slug")
	bar, err := fetchBar(barParam)
	if err != nil {
		fmt.Println("Error fetching bar:", err)
	}
	templates.Layout("Bar", templates.BarPage(bar)).Render(r.Context(), w)
}

func handleCreateBar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	decoder := form.NewDecoder()
	var userInput models.BarManual
	decoder.Decode(&userInput, r.PostForm)
	fmt.Println(r.PostForm)

	preview, err := database.InitiateCreateBar(&userInput)
	if err != nil {
		fmt.Println("Error generating preview:", err)
		http.Error(w, "Unable to create bar preview", http.StatusInternalServerError)
		return
	}
	templ.Handler(templates.BarPreview(preview)).ServeHTTP(w, r)
}

func fetchBar(lookupParam string) (models.Bar, error) {
	data, err := os.ReadFile("testdata.json")
	if err != nil {
		return models.Bar{}, fmt.Errorf("failed to read file: %w", err)
	}

	var bars []models.Bar
	if err := json.Unmarshal(data, &bars); err != nil {
		return models.Bar{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	for _, bar := range bars {
		if strings.EqualFold(bar.Slug, lookupParam) {
			return bar, nil
		}
	}

	return models.Bar{}, fmt.Errorf("no bar found for lookup: %s", lookupParam)
}
