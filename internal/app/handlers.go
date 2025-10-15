package app

import (
	"context"
	"fmt"
	"go-router/templates"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-playground/form"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (a *appl) handleCreateBar(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	decoder := form.NewDecoder()
	var userInput CreateBarForm
	decoder.Decode(&userInput, r.PostForm)
	log.Println(r.PostForm)

	err := CreateBarWithPrice(ctx, a.Pool, userInput)
	if err != nil {
		msg := fmt.Sprintf("Noe gikk galt: %s", err)
		templ.Handler(templates.Toast(msg)).ServeHTTP(w, r)
	} else {
		templ.Handler(templates.Toast("Bar opprettet!")).ServeHTTP(w, r)
	}
}
