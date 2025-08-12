package main

import (
	"fmt"
	"go-router/database"
	"go-router/internal/handlers"
	"go-router/internal/stores"
	"go-router/internal/utils"
	"go-router/models"
	"go-router/templates"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-playground/form"
)

func (a *app) handleHome(w http.ResponseWriter, r *http.Request) {
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

func (a *app) handleAbout(w http.ResponseWriter, r *http.Request) {
	data, err := database.GetAboutPageData(a.DB)
	if err != nil {
		fmt.Println("Error getting about info:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	templates.Layout("About", templates.About(data)).Render(r.Context(), w)
}

func (a *app) handleBar(w http.ResponseWriter, r *http.Request) {
	barParam := chi.URLParam(r, "slug")
	bar, err := database.GetBarBySlug(a.DB, barParam)

	if err != nil {
		fmt.Println("Error fetching bar:", err)
	}

	var user models.User
	c, err := r.Cookie("access_token")
	if err == nil && c.Value != "" {
		user.Admin = true
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

	brews := stores.AppStore.GetBreweriesData()

	templates.Layout("Bar", templates.BarPage(bar, hkeys, extra, brews, &user)).Render(r.Context(), w)
}

func (a *app) handleListFylke(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	params := map[string]string{
		"fylke": "/" + chi.URLParam(r, "fylke"),
	}
	nav, err := utils.SetNavParams(params)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig URL", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	navStore := models.Navigation{Level: "fylke", ID: nav.Fylke.ID}
	stores.SetNavData(sessionStore, sessID, navStore)

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

func (a *app) handleListKommune(w http.ResponseWriter, r *http.Request) {
	sessID := handlers.GetSessionID(r)
	params := map[string]string{
		"fylke":   "/" + chi.URLParam(r, "fylke"),
		"kommune": "/" + chi.URLParam(r, "fylke") + "/" + chi.URLParam(r, "kommune"),
	}
	nav, err := utils.SetNavParams(params)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.Layout("Ugyldig URL", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	navStore := models.Navigation{Level: "kommune", ID: nav.Kommune.ID}
	stores.SetNavData(sessionStore, sessID, navStore)

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

func (a *app) handleListSted(w http.ResponseWriter, r *http.Request) {
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

func (a *app) handleCustomTime(w http.ResponseWriter, r *http.Request) {
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

func (a *app) handleSearch(w http.ResponseWriter, r *http.Request) {
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

func (a *app) handleSearchBar(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	decoded, _ := url.QueryUnescape(searchTerm)

	result, err := database.GetBarSearchResult(a.DB, decoded)
	if err != nil {
		log.Fatalf("unable to get search result: %v", err)
	}

	templ.Handler(templates.BarSearchResult(result)).ServeHTTP(w, r)
}

func (a *app) handleFetchBar(w http.ResponseWriter, r *http.Request) {
	barSlug := r.URL.Query().Get("bar_slug")
	decoded, _ := url.QueryUnescape(barSlug)

	bar, err := database.GetBarBySlug(a.DB, decoded)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
		return
	}

	var hkeys []models.HappyKey
	if bar.TimedPrices {
		hkeys, err = database.GetHappyKeysByBarID(a.DB, bar.ID)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
			return
		}
	}

	var extra models.BarMetadata
	if bar.LinkedBar {
		extra, err = database.GetBarMetadata(a.DB, bar.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.Layout("Feil oppstod under behandling av data", templates.ErrorPage()).Render(r.Context(), w)
			return
		}
	}

	templ.Handler(templates.BarForm(bar, extra, hkeys)).ServeHTTP(w, r)
}
