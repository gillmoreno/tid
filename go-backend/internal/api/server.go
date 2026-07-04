package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"tid/go-backend/internal/config"
)

type App struct {
	cfg    config.Config
	logger *log.Logger
}

func NewApp(cfg config.Config, logger *log.Logger) *App {
	return &App{cfg: cfg, logger: logger}
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", a.handleHealth)
	r.Head("/health", a.handleHealth)

	r.Route("/api", func(api chi.Router) {
		api.Get("/site", a.handleSite)
		api.Get("/sections", a.handleSections)
		api.Get("/sections/{slug}", a.handleSection)
		api.Get("/items/{slug}", a.handleItem)
	})

	return r
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}