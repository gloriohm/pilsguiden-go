package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go-router/database"
	"go-router/internal/handlers"
	"go-router/models"
	"go-router/templates"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/form"
	"github.com/jackc/pgx/v5"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Database connected")

	bar, err := GetBarBySlug(ctx, conn, "cafe-sara")
	if err != nil {
		log.Fatalf("Unable to get bar: %v", err)
	}
	fmt.Println(bar.Name)

	bars, err := GetBarByFylke(ctx, conn, 2)
	if err != nil {
		log.Fatalf("Unable to get bars: %v", err)
	}
	fmt.Println(bars[0].Name)

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

	r.Route("/liste", func(r chi.Router) {
		r.Route("/{fylke}", func(r chi.Router) {
			r.Get("/", handlers.HandleListe)

			r.Route("/{kommune}", func(r chi.Router) {
				r.Get("/", handlers.HandleListe)

				r.Route("/{sted}", func(r chi.Router) {
					r.Get("/", handlers.HandleListe)
				})
			})
		})
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

func GetBarBySlug(ctx context.Context, conn *pgx.Conn, slug string) (*models.Bar, error) {
	query := `SELECT bar, size, pint FROM bars WHERE slug = $1`

	row := conn.QueryRow(ctx, query, slug)
	var bar models.Bar
	if err := row.Scan(&bar.Name, &bar.Size, &bar.Pint); err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}

	return &bar, nil
}

func GetBarByFylke(ctx context.Context, conn *pgx.Conn, fylke int) ([]models.Bar, error) {
	var bars []models.Bar
	query := `SELECT bar, size, current_pint FROM current_bars_view WHERE fylke = $1`

	rows, err := conn.Query(ctx, query, fylke)
	if err != nil {
		return bars, err
	}

	for rows.Next() {
		var bar models.Bar
		if err := rows.Scan(&bar.Name, &bar.Size, &bar.Pint); err != nil {
			return bars, fmt.Errorf("scanning row: %w", err)
		}
		bars = append(bars, bar)
	}

	if rows.Err() != nil {
		return bars, fmt.Errorf("iterating rows: %w", rows.Err())
	}

	return bars, nil
}
