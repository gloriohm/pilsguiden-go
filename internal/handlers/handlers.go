package handlers

import (
	"go-router/internal/stores"
	"net/http"

	"github.com/go-chi/chi"
)

func HandleListe(w http.ResponseWriter, r *http.Request) {
	fylke := chi.URLParam(r, "fylke")
	kommune := chi.URLParam(r, "kommune")
	sted := chi.URLParam(r, "sted")

	switch {
	case sted != "":
		stedID := stores.AppStore.GetStedBySlug(sted)
		if stedID == 0 {
			http.Error(w, "Ugyldig sted", http.StatusNotFound)
		}
	case kommune != "":
		kommuneID := stores.AppStore.GetKommuneBySlug(kommune)
		if kommuneID == 0 {
			http.Error(w, "Ugyldig sted", http.StatusNotFound)
		}
	case fylke != "":
		fylkeID := stores.AppStore.GetFylkeBySlug(fylke)
		if fylkeID == 0 {
			http.Error(w, "Ugyldig sted", http.StatusNotFound)
		}
	default:
		// invalid route
	}
}
