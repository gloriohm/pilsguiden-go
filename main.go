package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

type App struct {
	DB *pgx.Conn
}

var sessionStore = cache.New(30*time.Minute, 10*time.Minute)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	app := &App{DB: conn}
	database.InitStaticData(app.DB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(handlers.SessionMiddleware)
	r.Get("/", app.handleHome)
	r.Get("/about", app.handleAbout)
	r.Get("/bar/{slug}", app.handleBar)

	r.Route("/admin", func(r chi.Router) {
		r.Get("/create-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Create Bar", templates.BarManualForm()).Render(r.Context(), w)
		})
		r.Post("/create-bar", app.handleCreateBar)
		r.Get("/update-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Update Bar", templates.UpdateBar()).Render(r.Context(), w)
		})
	})

	r.Route("/liste", func(r chi.Router) {
		r.Post("/setCustomTime", app.handleCustomTime)
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

	// site-wide functionality endpoints
	r.Get("/search", app.handleSearch)
	r.Get("/search-bar", app.handleSearchBar)
	r.Get("/fetch-bar", app.handleFetchBar)

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

	var hkeys []models.HappyKey
	if bar.TimedPrices {
		hkeys, err = database.GetHappyKeysByBarID(a.DB, bar.ID)

		if err != nil {
			fmt.Println("Error fetching happy keys:", err)
		}
	}

	var extra models.BarMetadata
	if bar.LinkedBar {
		extra, err = database.GetBarMetadata(a.DB, bar.ID)
		if err != nil {
			fmt.Println("Error fetching bar metadata:", err)
		}
	}

	templates.Layout("Bar", templates.BarPage(bar, hkeys, extra)).Render(r.Context(), w)
}

func (a *App) handleCreateBar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	decoder := form.NewDecoder()
	var userInput models.BarManual
	decoder.Decode(&userInput, r.PostForm)
	fmt.Println(r.PostForm)

	err := database.CreateBar(a.DB, &userInput)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div class="p-4 bg-green-100 text-green-800 rounded">Noe gikk galt: %s</div>`, err)
	} else {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div class="p-4 bg-green-100 text-green-800 rounded">Bar created successfully!</div>`)
	}

}

func (a *App) handleListFylke(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	params := map[string]string{
		"fylke": "/" + chi.URLParam(r, "fylke"),
	}
	nav, err := utils.SetNavParams(params)
	navStore := models.Navigation{Level: "fylke", ID: nav.Fylke.ID}
	stores.SetNavData(sessionStore, sessID, navStore)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig URL", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	var bars []models.Bar
	pref := stores.GetSessionData(sessionStore, sessID)
	if pref.Preferences.CustomTime {
		bars, err = database.GetBarsByLocationAndTime(a.DB, nav.Fylke.ID, "fylke", pref.Preferences.Date, pref.Preferences.Time.Format("15:04:05"))
	} else {
		bars, err = database.GetBarsByLocation(a.DB, nav.Fylke.ID, "fylke")
	}
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	nextLocations := utils.ExtractSortedUniqueKommuner(bars)
	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListKommune(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	params := map[string]string{
		"fylke":   "/" + chi.URLParam(r, "fylke"),
		"kommune": "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune"),
	}
	nav, err := utils.SetNavParams(params)
	navStore := models.Navigation{Level: "kommune", ID: nav.Kommune.ID}
	stores.SetNavData(sessionStore, sessID, navStore)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig URL", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	var bars []models.Bar
	pref := stores.GetSessionData(sessionStore, sessID)
	if pref.Preferences.CustomTime {
		bars, err = database.GetBarsByLocationAndTime(a.DB, nav.Kommune.ID, "kommune", pref.Preferences.Date, pref.Preferences.Time.Format("15:04:05"))
	} else {
		bars, err = database.GetBarsByLocation(a.DB, nav.Kommune.ID, "kommune")
	}
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	nextLocations := utils.ExtractSortedUniqueSteder(bars)
	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleListSted(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	params := map[string]string{
		"fylke":   "/" + chi.URLParam(r, "fylke"),
		"kommune": "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune"),
		"sted":    "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune") + "/" + chi.URLParam(r, "sted"),
	}
	nav, err := utils.SetNavParams(params)
	navStore := models.Navigation{Level: "sted", ID: nav.Sted.ID}
	stores.SetNavData(sessionStore, sessID, navStore)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig URL", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	var bars []models.Bar
	pref := stores.GetSessionData(sessionStore, sessID)
	if pref.Preferences.CustomTime {
		bars, err = database.GetBarsByLocationAndTime(a.DB, nav.Sted.ID, "sted", pref.Preferences.Date, pref.Preferences.Time.Format("15:04:05"))
	} else {
		bars, err = database.GetBarsByLocation(a.DB, nav.Sted.ID, "sted")
	}
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	nextLocations := utils.ExtractSortedUniqueSteder(bars)

	templates.Layout("List", templates.ListLayout(templates.NavTree(nav), templates.LocationLinks(nextLocations), templates.List(bars))).Render(r.Context(), w)
}

func (a *App) handleCustomTime(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
		return
	}
	decoder := form.NewDecoder()

	var userInput models.RawCustomTime
	decoder.Decode(&userInput, r.PostForm)

	// get closest date from day
	customDate := stores.GetClosestDate(userInput.Day)

	// parse string to timestamp
	t, _ := time.Parse("15:04", userInput.Time)
	pref := models.Preferences{
		CustomTime: true,
		Time:       t,
		Date:       customDate,
	}

	stores.SetSessionPrefs(sessionStore, sessID, pref)

	sessData := stores.GetSessionData(sessionStore, sessID)
	fmt.Println(sessData.Navigation.Level)

	bars, err := database.GetBarsByLocationAndTime(a.DB, sessData.Navigation.ID, sessData.Navigation.Level, sessData.Preferences.Date, sessData.Preferences.Time.Format("15:04:05"))
	if err != nil {
		log.Fatalf("unable to get bars: %v", err)
	}
	templ.Handler(templates.List(bars)).ServeHTTP(w, r)
}

func (a *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	decoded, _ := url.QueryUnescape(searchTerm)
	if len(decoded) > 2 {
		result, err := database.GetSearchResult(a.DB, decoded)
		if err != nil {
			log.Fatalf("unable to get search result: %v", err)
		}

		templ.Handler(templates.SearchResult(result)).ServeHTTP(w, r)
	} else {
		templ.Handler(templates.SearchResult([]models.SearchResult{})).ServeHTTP(w, r)
	}

}

func (a *App) handleSearchBar(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	decoded, _ := url.QueryUnescape(searchTerm)

	result, err := database.GetBarSearchResult(a.DB, decoded)
	if err != nil {
		log.Fatalf("unable to get search result: %v", err)
	}

	templ.Handler(templates.BarSearchResult(result)).ServeHTTP(w, r)
}

func (a *App) handleFetchBar(w http.ResponseWriter, r *http.Request) {
	barSlug := r.URL.Query().Get("bar_slug")
	decoded, _ := url.QueryUnescape(barSlug)

	bar, err := database.GetBarBySlug(a.DB, decoded)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
	}

	var hkeys []models.HappyKey
	if bar.TimedPrices {
		hkeys, err = database.GetHappyKeysByBarID(a.DB, bar.ID)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
		}
	}

	var extra models.BarMetadata
	if bar.LinkedBar {
		extra, err = database.GetBarMetadata(a.DB, bar.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
		}
	}

	templ.Handler(templates.BarForm(bar, extra, hkeys)).ServeHTTP(w, r)
}
