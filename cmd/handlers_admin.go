package main

import (
	"context"
	"go-router/database"
	"go-router/templates"
	"log"
	"net/http"
)

func (a *app) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	pendingPrices, err := database.GetPendingPrices(ctx, a.Pool)
	if err != nil {
		log.Printf("unable to get prices: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		generateErrorPage(w, r.Context(), "Bad request", true)
	}

	templates.Layout("Godkjenn prisoppdateringer", true, templates.Dashboard(pendingPrices)).Render(r.Context(), w)
}
