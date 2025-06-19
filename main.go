package main

import (
	"fmt"
	"log"
	"net/http"

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

type App struct {
	DB *pgx.Conn
}

func main() {
	conn, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	app := &App{DB: conn}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout("Home", templates.Home()).Render(r.Context(), w)
	})
	r.Get("/about", app.handleAbout)
	r.Get("/bar/{slug}", app.handleBar)

	r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		bars, err := database.GetBarsByFylke(conn, 2)
		if err != nil {
			log.Fatalf("unable to get bars: %v", err)
		}
		templates.Layout("List", templates.List(bars)).Render(r.Context(), w)
	})

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

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {

}

func (a *App) handleAbout(w http.ResponseWriter, r *http.Request) {
	data, err := database.GetAboutPageData(a.DB)
	if err != nil {
		fmt.Println("Error getting about info:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	templates.Layout("About", templates.About(data)).Render(r.Context(), w)
}

func (a *App) handleBar(w http.ResponseWriter, r *http.Request) {
	barParam := chi.URLParam(r, "slug")
	bar, err := database.GetBarBySlug(a.DB, barParam)
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
