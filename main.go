package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-router/database"
	"go-router/internal/handlers"
	"go-router/internal/stores"
	"go-router/internal/utils"
	"go-router/models"
	"go-router/templates"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/form"
	"github.com/jackc/pgx/v5"
	"github.com/patrickmn/go-cache"
)

type App struct {
	DB *pgx.Conn
}

var sessionStore = cache.New(30*time.Minute, 10*time.Minute)

func main() {
	conn, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	app := &App{DB: conn}
	database.InitStaticData(app.DB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(handlers.SessionMiddleware(sessionStore))
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

	// r.Get("/kart", app.handleMap)

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
	val := handlers.GetSessionID(r)
	fmt.Println(val)
	params := map[string]string{
		"fylke": "/" + chi.URLParam(r, "fylke"),
	}
	nav, err := utils.SetNavParams(params)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig sted", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	nextLocations := stores.AppStore.GetLocationsByParent(nav.Fylke.ID, "kommune") // get all kommuner under current fylke
	bars, err := database.GetBarsByLocation(a.DB, nav.Fylke.ID, "fylke")
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}

	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListKommune(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"fylke":   "/" + chi.URLParam(r, "fylke"),
		"kommune": "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune"),
	}
	nav, err := utils.SetNavParams(params)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig sted", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	nextLocations := stores.AppStore.GetLocationsByParent(nav.Kommune.ID, "sted") //get all steder under current kommune

	bars, err := database.GetBarsByLocation(a.DB, nav.Kommune.ID, "sted") //bytt til kommune ved migrering
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}

	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListSted(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"fylke":   "/" + chi.URLParam(r, "fylke"),
		"kommune": "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune"),
		"sted":    "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune") + "/" + chi.URLParam(r, "sted"),
	}
	nav, err := utils.SetNavParams(params)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig sted", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	nextLocations := stores.AppStore.GetLocationsByParent(nav.Kommune.ID, "sted") //get all steder under current kommune since we go no deeper

	bars, err := database.GetBarsByLocation(a.DB, nav.Sted.ID, "nabolag") //bytt til sted ved migrering
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}

	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}
