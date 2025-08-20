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
			templates.Layout("Vær hilset, Admin", true, templates.Search()).Render(r.Context(), w)
		})
		r.Get("/create-bar", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Opprett bar", true, templates.BarManualForm()).Render(r.Context(), w)
		})
		r.Get("/update-bar/{id}", app.handleUpdateBarForm)
		r.Get("/create-brewery", func(w http.ResponseWriter, r *http.Request) {
			templates.Layout("Opprett bryggeri", true, templates.CreateBrewery()).Render(r.Context(), w)
		})
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout("Login", true, templates.Login()).Render(r.Context(), w)
	})

	r.Post("/auth/login", auth.LoginHandler)

	r.Route("/liste", func(r chi.Router) {
		r.Post("/setCustomTime", app.handleCustomTime)
		r.Route("/{fylke}", func(r chi.Router) {
			r.Get("/", app.handleList)

			r.Route("/{kommune}", func(r chi.Router) {
				r.Get("/", app.handleList)

				r.Route("/{sted}", func(r chi.Router) {
					r.Get("/", app.handleList)
				})
			})
		})
	})

	// r.Get("/kart", app.handleMap)

	// site-wide functionality endpoints
	r.Get("/search", app.handleSearch)
	r.Route("/logic", func(r chi.Router) {
		r.Get("/brewery-picker", handleBreweryPicker)
		r.Get("/consent-updater", handleConsentUpdater)
		r.Get("/price-confirmer", handlePriceConfirmer)
		r.Get("/price-updater", handlePriceUpdater)
		r.Get("/do-nothing", handleDoNothing)
		r.Post("/confirm-price", app.handleConfirmPrice)
		r.Post("/set-consent", handleConsent)
		r.Post("/update-bar", app.handleUpdateBar)
		r.Post("/update-brewery", app.handleUpdateBrewery)
		r.Post("/update-price", app.handleUpdatePrice)
		r.Post("/test-toast", testToast)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(app.APIKeyMiddleware)
			r.Post("/create-brewery", app.handleCreateBrewery)
			r.Post("/create-bar", app.handleCreateBar)
			r.Get("/fetch-bars", app.handleFetchBars)
		})
	})

	return r
}
