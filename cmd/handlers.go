package main

import (
	"fmt"
	"go-router/database"
	"go-router/internal/auth"
	"go-router/internal/handlers"
	"go-router/internal/stores"
	"go-router/internal/utils"
	"go-router/models"
	"go-router/templates"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"github.com/go-playground/form"
)

func (a *app) handleHome(w http.ResponseWriter, r *http.Request) {
	sessionData := handlers.GetSessionData(r.Context())
	fylker := stores.AppStore.GetFylkerData()
	baseFylke := utils.ToBase(fylker)
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
	templates.Layout("Pilsguiden", sessionData.Consented, templates.Home(totalBars, baseFylke, topTen, bottomTen)).Render(r.Context(), w)
}

func (a *app) handleAbout(w http.ResponseWriter, r *http.Request) {
	sessionData := handlers.GetSessionData(r.Context())
	data, err := database.GetAboutPageData(a.DB)
	if err != nil {
		fmt.Println("Error getting about info:", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
	}
	templates.Layout("Om Pilsguiden", sessionData.Consented, templates.About(data)).Render(r.Context(), w)
}

func (a *app) handleBar(w http.ResponseWriter, r *http.Request) {
	sessionData := handlers.GetSessionData(r.Context())
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

	templates.Layout(bar.Name, sessionData.Consented, templates.BarPage(bar, hkeys, extra, &user)).Render(r.Context(), w)
}

func (a *app) handleList(w http.ResponseWriter, r *http.Request) {
	sessionData := handlers.GetSessionData(r.Context())
	sessID := sessionData.SessionID
	consented := sessionData.Consented
	fylke := chi.URLParam(r, "fylke")
	kommune := chi.URLParam(r, "kommune")
	sted := chi.URLParam(r, "sted")

	nav, current, err := setNavParams(fylke, kommune, sted)
	if err != nil {
		log.Printf("Error setting nav: %s \n", err)
		w.WriteHeader(http.StatusNotFound)
		generateErrorPage(w, r.Context(), "Ugyldig URL", sessionData.Consented)
		return
	}

	navStore := models.Navigation{Level: current.Name, ID: current.ID}
	stores.SetNavData(sessionStore, sessID, navStore)

	var bars []models.BarView
	pref := stores.GetSessionData(sessionStore, sessID)
	if pref.Preferences.CustomTime {
		bars, err = database.GetBarsByLocationAndTime(a.DB, current.ID, current.Name, pref.Preferences.Date, pref.Preferences.Time.Format("15:04:05"))
	} else {
		bars, err = database.GetBarsByLocation(a.DB, current.ID, current.Name)
	}
	if err != nil {
		log.Printf("unable to get bars: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", sessionData.Consented)
	}

	nextLocations := extractNextLocs(bars, current.Name)
	templates.Layout("List", consented, templates.ListView(nav, nextLocations, bars)).Render(r.Context(), w)
}

func (a *app) handleCustomTime(w http.ResponseWriter, r *http.Request) {
	sessionData := handlers.GetSessionData(r.Context())
	sessID := sessionData.SessionID
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", sessionData.Consented)
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
		log.Printf("unable to get bars: %v", err)
	}
	templ.Handler(templates.List(bars)).ServeHTTP(w, r)
}

func (a *app) handleSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search")
	decoded, _ := url.QueryUnescape(searchTerm)
	if len(decoded) > 2 {
		result, err := database.GetSearchResult(a.DB, decoded)
		if err != nil {
			log.Printf("unable to get search result: %v", err)
		}

		templ.Handler(templates.SearchResult(result)).ServeHTTP(w, r)
	} else {
		templ.Handler(templates.SearchResult([]models.SearchResult{})).ServeHTTP(w, r)
	}
}

func (a *app) handleUpdateBarForm(w http.ResponseWriter, r *http.Request) {
	barParam := chi.URLParam(r, "id")
	barID, err := strconv.Atoi(barParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Ugyldig bar-ID", true)
		return
	}

	v := r.Context().Value(auth.UserIDKey)
	user, ok := v.(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Ugyldig user-ID", true)
		return
	}

	bar, err := database.GetBarByID(a.DB, barID)
	if err != nil {
		log.Printf("unable to fetch bar by ID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Ugyldig bar-ID", true)
		return
	}

	var hkeys []models.HappyKey
	if bar.TimedPrices {
		hkeys, err = database.GetHappyKeysByBarID(a.DB, barID)
		if err != nil {
			log.Printf("unable to fetch hkeys for bar with ID %d: %v", barID, err)
		}
	}

	stores.SetUpdateBarStore(sessionStore, user, models.UpdateBarStore{BarID: bar.ID, Price: bar.Price, Size: bar.Size})

	brews := stores.AppStore.GetBreweriesData()

	templates.Layout("Oppdater bar", true, templates.UpdateBarForm(bar, hkeys, brews)).Render(r.Context(), w)
}

func (a *app) handleUpdateBar(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(auth.UserIDKey)
	user, ok := v.(string)
	if !ok {
		log.Println("error getting user ID")
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Ugyldig user-ID", true)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", true)
		return
	}
	decoder := form.NewDecoder()

	var updatedBar models.BarUpdateForm
	decoder.Decode(&updatedBar, r.PostForm)
	fmt.Println(updatedBar)

	compare := stores.GetUpdateBarStore(sessionStore, user)

	if compare.BarID != updatedBar.ID {
		log.Println("Bar id does not match bar id in store")
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", true)
		return
	}
	if compare.Price != updatedBar.Price || compare.Size != updatedBar.Size {
		newPrice := models.Price{
			BarID:        updatedBar.ID,
			Price:        updatedBar.Price,
			Size:         updatedBar.Size,
			Pint:         database.ToPint(updatedBar.Price, updatedBar.Size),
			PriceUpdated: time.Now(),
			PriceChecked: time.Now(),
		}
		if err := database.UpdateCurrentAndHistoricPrice(a.DB, newPrice); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			generateErrorPage(w, r.Context(), "Bad request", true)
			return
		}
	}

	// run update other info logic
	if err := database.UpdateBarData(a.DB, updatedBar); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", true)
		return
	}

	w.WriteHeader(http.StatusOK)
	templ.Handler(templates.Toast("Bar oppdatert!")).ServeHTTP(w, r)
}
