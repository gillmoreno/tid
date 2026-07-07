package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"tid/go-backend/internal/config"
	"tid/go-backend/internal/db"
	"tid/go-backend/internal/factory"
)

type App struct {
	cfg     config.Config
	logger  *log.Logger
	factory *factory.Store
	runner  factory.Runner
}

func NewApp(cfg config.Config, logger *log.Logger) (*App, error) {
	provider, err := db.NewProvider(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	store := factory.NewStore(provider)
	if err := seedFactoryStore(store, cfg.RepoRoot); err != nil {
		logger.Printf("factory seed warning: %v", err)
	}
	if err := store.SeedIdeas(); err != nil {
		logger.Printf("ideas seed warning: %v", err)
	}
	app := &App{
		cfg:     cfg,
		logger:  logger,
		factory: store,
		runner:  factory.NewRunner(cfg.RepoRoot),
	}
	return app, nil
}

func (a *App) Close() error {
	return nil
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
		a.mountFactoryRoutes(api)
	})

	return r
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}