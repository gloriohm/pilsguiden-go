package main

import (
	"go-router/internal/auth"
	"go-router/internal/handlers"
	"go-router/templates"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *app) routes() http.Handler {
	fileServer := http.FileServer(http.Dir("./static"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(handlers.SessionMiddleware)
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	r.Get("/", app.handleHome)
	r.Get("/om-oss", app.handleAbout)
	r.Get("/bar/{slug}", app.handleBar)

	r.Route("/admin", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Vær hilset, Admin", templates.Search()).Render(r.Context(), w)
		})
		r.Get("/create-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Opprett bar", templates.BarManualForm()).Render(r.Context(), w)
		})
		r.Get("/update-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Oppdater Bar", templates.UpdateBar()).Render(r.Context(), w)
		})
		r.Get("/create-brewery", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Opprett bryggeri", templates.CreateBrewery()).Render(r.Context(), w)
		})
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout("Login", templates.Login()).Render(r.Context(), w)
	})

	r.Post("/auth/login", auth.LoginHandler)

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

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(app.APIKeyMiddleware)
			r.Post("/update-brewery", app.handleUpdateBrewery)
			r.Post("/create-brewery", app.handleCreateBrewery)
			r.Post("/create-bar", app.handleCreateBar)
			r.Post("/confirm-price", app.handleConfirmPrice)
			r.Get("/fetch-bars", app.handleFetchBars)
		})
	})

	return r
}
