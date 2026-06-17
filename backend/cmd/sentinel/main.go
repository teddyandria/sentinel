// Command sentinel est le point d'entrée de l'application : il câble les couches
// et gère l'arrêt propre (graceful shutdown).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teddyandria/sentinel/internal/api"
	"github.com/teddyandria/sentinel/internal/config"
	"github.com/teddyandria/sentinel/internal/domain"
	"github.com/teddyandria/sentinel/internal/fetcher"
	"github.com/teddyandria/sentinel/internal/geocoder"
	"github.com/teddyandria/sentinel/internal/pipeline"
	"github.com/teddyandria/sentinel/internal/scheduler"
	"github.com/teddyandria/sentinel/internal/storage"
	"github.com/teddyandria/sentinel/pkg/newsapi"
)

func main() {
	if err := run(); err != nil {
		slog.Error("démarrage impossible", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := newLogger(cfg.LogLevel)

	// Contexte annulé sur SIGINT/SIGTERM : c'est ce qui déclenche le graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := storage.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()

	// Une source par topic : chaque NewsAPIFetcher ne récupère et ne tague que son sujet.
	newsClient := newsapi.New(cfg.NewsAPIKey)
	fetchers := make([]fetcher.Fetcher, 0, len(domain.AllowedTopics))
	for _, topic := range domain.AllowedTopics {
		fetchers = append(fetchers, fetcher.NewNewsAPIFetcher(newsClient, topic, log))
	}
	geo := geocoder.NewStaticGeocoder(geocoder.DefaultCities())

	// Le pipeline (fetch -> geocode -> store) est déclenché périodiquement par le scheduler.
	pipe := pipeline.New(fetchers, geo, store, log)
	sched := scheduler.New(cfg.FetchInterval, pipe.Run, log)
	go sched.Start(ctx)

	apiServer := api.NewServer(store, cfg.MapboxToken, log)
	httpSrv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apiServer.Routes(cfg.WebDir),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		url := "http://localhost" + cfg.HTTPAddr
		log.Info("serveur HTTP démarré — carte disponible sur "+url, "addr", cfg.HTTPAddr)
		// ListenAndServe renvoie ErrServerClosed lors d'un Shutdown : ce n'est pas une erreur.
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info("signal d'arrêt reçu, fermeture en cours...")
	}

	// On laisse 10s aux requêtes en cours de se terminer avant de couper.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("arrêt propre terminé")
	return nil
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
