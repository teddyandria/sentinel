// Package api expose les articles via HTTP et sert le frontend.
package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/teddyandria/sentinel/internal/storage"
)

type Server struct {
	store       storage.Store
	mapboxToken string // exposé au frontend via /api/config
	log         *slog.Logger
}

func NewServer(store storage.Store, mapboxToken string, log *slog.Logger) *Server {
	return &Server{store: store, mapboxToken: mapboxToken, log: log}
}

// Routes assemble l'API JSON (/api/*) et le service des fichiers statiques du frontend.
func (s *Server) Routes(webDir string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer) // transforme un panic en réponse 500 propre

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/config", s.handleConfig)
		r.Get("/topics", s.handleTopics)
		r.Get("/articles", s.handleListArticles)
	})

	fileServer := http.FileServer(http.Dir(webDir))
	r.Handle("/*", fileServer)

	return r
}
