package main

import (
	"go-router/templates"
	"net/http"

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

func handleUpdateConsent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	templ.Handler(templates.UpdateConsent()).ServeHTTP(w, r)
}
