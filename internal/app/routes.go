package app

import (
	"context"
	"go-router/internal/bars"
	"go-router/internal/handlers"
	"go-router/internal/stores"
	"go-router/templates"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi"
	"golang.org/x/sync/errgroup"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func NewRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Home)
	r.Get("/om-oss", h.About)
	r.Get("/kontakt", h.Contact)
	return r
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	sessionData := handlers.GetSessionData(ctx)
	data, err := h.svc.LoadHomeData(ctx)

	if err != nil {
		log.Printf("home: data load error: %v", err)
		http.Error(w, "Feil under lasting av data", http.StatusInternalServerError)
		return
	}

	c := templates.Layout("Pilsguiden", sessionData.Consented, templates.Home(data))
	templ.Handler(c).ServeHTTP(w, r)
}

func (s *Service) LoadHomeData(ctx context.Context) (HomeData, error) {
	var eg errgroup.Group
	data := HomeData{}

	eg.Go(func() error {
		n, err := bars.GetTotalBars(ctx, a.Pool)
		if err == nil {
			data.TotalBars = n
		}
		return err
	})
	eg.Go(func() error {
		t, err := bars.GetTopTen(ctx, a.Pool)
		if err == nil {
			data.TopTenBars = t
		}
		return err
	})
	eg.Go(func() error {
		b, err := bars.GetBottomTen(ctx, a.Pool)
		if err == nil {
			data.BottomTenBars = b
		}
		return err
	})

	err := eg.Wait()
	if err != nil {
		return data, err
	}

	fylker := stores.AppStore.GetFylkerData()
	baseFylke := bars.ToBase(fylker)
	data.Fylker = baseFylke

	return data, nil
}
