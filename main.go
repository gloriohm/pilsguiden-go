package main

import (
	"fmt"
	"log"
	"net/http"

	"go-router/database"
	"go-router/internal/stores"
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
	database.InitStaticData(app.DB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", app.handleHome)
	r.Get("/about", app.handleAbout)
	r.Get("/bar/{slug}", app.handleBar)

	r.Route("/admin", func(r chi.Router) {
		r.Get("/create-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Create Bar", templates.BarManualForm()).Render(r.Context(), w)
		})
		r.Post("/fetch-osm", handleCreateBar)
	})

	r.Route("/liste", func(r chi.Router) {
		r.Route("/{fylke}", func(r chi.Router) {
			r.Get("/", app.handleListFylke)

			r.Route("/{kommune}", func(r chi.Router) {
				r.Get("/", app.handleListKommune)

				r.Route("/{sted}", func(r chi.Router) {
					r.Get("/", app.handleListSted)
				})
			})
		})
	})

	http.ListenAndServe(":3000", r)
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	fylker := stores.AppStore.GetFylkerData()
	totalBars, err := database.GetTotalBars(a.DB)
	if err != nil || totalBars == 0 {
		fmt.Println("Error total bars:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	topTen, err := database.GetTopTen(a.DB)
	if err != nil {
		fmt.Println("Error top ten bars:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	bottomTen, err := database.GetBottomTen(a.DB)
	if err != nil {
		fmt.Println("Error loading bottom ten bars:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	templates.Layout("Home", templates.Home(totalBars, fylker, templates.List(topTen), templates.List(bottomTen))).Render(r.Context(), w)
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

func (a *App) handleListFylke(w http.ResponseWriter, r *http.Request) {
	currentLocation := "/" + chi.URLParam(r, "fylke")
	locationID := stores.AppStore.GetLocationBySlug(currentLocation, "fylke")
	nextLocations := stores.AppStore.GetLocationsByParent(locationID, "kommune")
	if locationID == 0 {
		http.Error(w, "Ugyldig sted", http.StatusNotFound)
	}
	bars, err := database.GetBarsByFylke(a.DB, locationID)
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	var urlParts []models.UrlPair
	urlParts = append(urlParts, models.UrlPair{
		Name: "Oslo",
		Slug: "oslo",
	})

	templates.Layout("List", templates.ListLayout(templates.NavTree(urlParts), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListKommune(w http.ResponseWriter, r *http.Request) {
	currentLocation := "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune")
	locationID := stores.AppStore.GetLocationBySlug(currentLocation, "kommune")
	nextLocations := stores.AppStore.GetLocationsByParent(locationID, "sted")
	if locationID == 0 {
		http.Error(w, "Ugyldig sted", http.StatusNotFound)
	}
	bars, err := database.GetBarsByKommune(a.DB, locationID)
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	var urlParts []models.UrlPair
	urlParts = append(urlParts, models.UrlPair{
		Name: "Oslo",
		Slug: "oslo",
	})

	templates.Layout("List", templates.ListLayout(templates.NavTree(urlParts), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListSted(w http.ResponseWriter, r *http.Request) {
	currentLocation := "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune") + "/" + chi.URLParam(r, "sted")
	parentKommune := "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune")
	locationID := stores.AppStore.GetLocationBySlug(currentLocation, "sted")
	parentID := stores.AppStore.GetLocationBySlug(parentKommune, "kommune")
	nextLocations := stores.AppStore.GetLocationsByParent(parentID, "sted")
	if locationID == 0 {
		http.Error(w, "Ugyldig sted", http.StatusNotFound)
	}
	bars, err := database.GetBarsBySted(a.DB, locationID)
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	var urlParts []models.UrlPair
	urlParts = append(urlParts, models.UrlPair{
		Name: "Oslo",
		Slug: "oslo",
	})

	templates.Layout("List", templates.ListLayout(templates.NavTree(urlParts), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}
