package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
)

func HandleListe(w http.ResponseWriter, r *http.Request) {
	fylke := chi.URLParam(r, "fylke")
	kommune := chi.URLParam(r, "kommune")
	sted := chi.URLParam(r, "sted")

	switch {
	case sted != "":
		// list for sted
	case kommune != "":
		// list for kommune
	case fylke != "":
		// list for fylke
	default:
		// invalid route
	}
}
