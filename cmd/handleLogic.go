package main

import (
	"fmt"
	"go-router/database"
	"go-router/internal/stores"
	"go-router/models"
	"go-router/templates"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
)

func handleConsent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	level := r.FormValue("level")

	http.SetCookie(w, &http.Cookie{
		Name:     "user_consent",
		Value:    level,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

func handleDoNothing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleConsentUpdater(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	templ.Handler(templates.ConsentUpdater()).ServeHTTP(w, r)
}

func handleBreweryPicker(w http.ResponseWriter, r *http.Request) {
	barID := r.URL.Query().Get("id")
	brews := stores.AppStore.GetBreweriesData()
	w.WriteHeader(http.StatusOK)
	templ.Handler(templates.BreweryPicker(barID, brews)).ServeHTTP(w, r)
}

func (a *app) handleUpdateBrewery(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}
	rawBar := r.FormValue("bar_id")
	barID, err := strconv.Atoi(rawBar)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	brewery := r.FormValue("new_brew")

	err = database.UpdateBreweryWhereUnknown(a.DB, barID, brewery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	templ.Handler(templates.UpdateInterface("Bryggeri oppdatert!", brewery, "brewery")).ServeHTTP(w, r)
}

func testToast(w http.ResponseWriter, r *http.Request) {
	templ.Handler(templates.Toast("fjern meg!")).ServeHTTP(w, r)
}

func handlePriceConfirmer(w http.ResponseWriter, r *http.Request) {
	idType := r.URL.Query().Get("type")
	id := r.URL.Query().Get("id")

	templ.Handler(templates.PriceConfirmer(idType, id)).ServeHTTP(w, r)
}

func handlePriceUpdater(w http.ResponseWriter, r *http.Request) {
	idType := r.URL.Query().Get("type")
	id := r.URL.Query().Get("id")
	size := r.URL.Query().Get("size")

	templ.Handler(templates.PriceUpdater(idType, id, size)).ServeHTTP(w, r)
}

func (a *app) handleConfirmPrice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}
	target := r.FormValue("update_target")
	parts := strings.Split(target, ":")
	idType := parts[0]
	rawID := parts[1]
	id, err := strconv.Atoi(rawID)

	if err != nil {
		http.Error(w, "Id not of type int", http.StatusBadRequest)
		return
	}

	timestamp := time.Now()

	var targetElementID string
	switch idType {
	case "bar":
		targetElementID = "regular_checked"
		err = database.UpdatePriceChecked(a.DB, "bars", timestamp, id)
		if err != nil {
			log.Print(err)
			http.Error(w, "Not able to confirm price", http.StatusBadRequest)
			return
		}
	case "hkey":
		targetElementID = fmt.Sprintf("happy_checked_%s", rawID)
		err = database.UpdatePriceChecked(a.DB, "happy_keys", timestamp, id)
		if err != nil {
			log.Print(err)
			http.Error(w, "Not able to confirm price", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Not of type bar or hkey", http.StatusBadRequest)
		return
	}

	timeString := templates.FormatNorwegianDate(timestamp)

	templ.Handler(templates.UpdateInterface("Pris bekreftet! \nTakk for at du bidrar 🍻", timeString, targetElementID)).ServeHTTP(w, r)
}

func (a *app) handleUpdatePrice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}
	target := r.FormValue("update_target")
	parts := strings.Split(target, ":")
	idType := parts[0]
	rawID := parts[1]
	id, err := strconv.Atoi(rawID)

	if err != nil {
		http.Error(w, "Id not of type int", http.StatusBadRequest)
		return
	}

	rawNewPrice := r.FormValue("new_price")
	newPrice, err := strconv.Atoi(rawNewPrice)
	if err != nil {
		http.Error(w, "Id not of type int", http.StatusBadRequest)
		return
	}

	sizeStr := r.FormValue("size")
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		http.Error(w, "invalid size", http.StatusBadRequest)
		return
	}

	newPint := database.ToPint(newPrice, size)

	timestamp := time.Now()

	payload := models.UpdatedPrice{
		TargetID:     id,
		TargetTable:  idType,
		Price:        newPrice,
		Size:         size,
		Pint:         newPint,
		PriceUpdated: timestamp,
		PriceChecked: timestamp,
	}

	err = database.UpdatePricePublic(a.DB, payload)
	if err != nil {
		log.Print(err)
		http.Error(w, "not able to update price", http.StatusBadRequest)
		return
	}

	templ.Handler(templates.Toast("Pris sendt til kontroll! \n Takk for at du bidrar 🍻")).ServeHTTP(w, r)
}
