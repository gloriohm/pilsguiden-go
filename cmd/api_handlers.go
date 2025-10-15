package main

import (
	"context"
	"encoding/json"
	"go-router/database"
	"go-router/internal/stores"
	"go-router/templates"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

func (a *appl) handleCreateBrewery(w http.ResponseWriter, r *http.Request) {
	newBrew := r.FormValue("new_brew")

	if newBrew == "" {
		templ.Handler(templates.Toast("Navnet må inneholde minst én bokstav.")).ServeHTTP(w, r)
		return
	}

	exists := stores.AppStore.BreweryInBreweries(newBrew)

	if exists {
		templ.Handler(templates.Toast("Bryggeri finnes fra før.")).ServeHTTP(w, r)
		return
	}

	err := database.CreateBrewery(a.Pool, newBrew)
	if err != nil {
		http.Error(w, "could not create brewery", http.StatusBadRequest)
		return
	}

	breweries, err := database.GetBreweries(a.Pool)
	if err != nil {
		log.Fatalf("failed to load breweries: %v", err)
	}

	stores.AppStore.UpdateBreweries(breweries)

	templ.Handler(templates.Toast("Bryggeri opprettet!")).ServeHTTP(w, r)
}

func (a *appl) handleFetchBars(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

}

func (app *appl) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing API key"})
			return
		}

		var customerKey string
		var active bool
		err := app.Pool.QueryRow(r.Context(),
			"SELECT key, active FROM api_keys WHERE key=$1", apiKey,
		).Scan(&customerKey, &active)

		if err != nil || !active {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid API key"})
			return
		}

		type apiKeyContext string
		const customerIDKey apiKeyContext = "customer_key"

		ctx := context.WithValue(r.Context(), customerIDKey, customerKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
