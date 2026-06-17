// Package scheduler exécute un Job à intervalle régulier.
package scheduler

import (
	"context"
	"log/slog"
	"time"
)

type Job func(ctx context.Context) error

type Scheduler struct {
	interval time.Duration
	job      Job
	log      *slog.Logger
}

func New(interval time.Duration, job Job, log *slog.Logger) *Scheduler {
	return &Scheduler{interval: interval, job: job, log: log}
}

// Start est bloquant (à lancer dans une goroutine) et s'arrête quand ctx est annulé.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.runOnce(ctx) // exécution immédiate, sans attendre le premier tick

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler arrêté")
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

// runOnce exécute le job en loguant l'erreur sans interrompre la boucle.
func (s *Scheduler) runOnce(ctx context.Context) {
	if err := s.job(ctx); err != nil {
		s.log.Error("job du scheduler échoué", "err", err)
	}
}
