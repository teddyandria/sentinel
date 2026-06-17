// Package pipeline orchestre l'ingestion : fetch -> geocode -> store.
package pipeline

import (
	"context"
	"log/slog"

	"github.com/teddyandria/sentinel/internal/fetcher"
	"github.com/teddyandria/sentinel/internal/geocoder"
	"github.com/teddyandria/sentinel/internal/storage"
)

type Pipeline struct {
	fetcher  fetcher.Fetcher
	geocoder geocoder.Geocoder
	store    storage.Store
	log      *slog.Logger
}

func New(f fetcher.Fetcher, g geocoder.Geocoder, s storage.Store, log *slog.Logger) *Pipeline {
	return &Pipeline{fetcher: f, geocoder: g, store: s, log: log}
}

// Run exécute une passe d'ingestion. Une erreur sur un article n'interrompt pas les autres.
func (p *Pipeline) Run(ctx context.Context) error {
	articles, err := p.fetcher.Fetch(ctx)
	if err != nil {
		return err
	}

	var saved, geocoded int
	for _, a := range articles {
		loc, err := p.geocoder.Geocode(ctx, a.Title+" "+a.Description)
		if err != nil {
			p.log.Warn("géocodage échoué", "url", a.URL, "err", err)
		} else if loc != nil {
			a.Location = loc
			geocoded++
		}

		if err := p.store.Save(ctx, a); err != nil {
			p.log.Error("sauvegarde échouée", "url", a.URL, "err", err)
			continue
		}
		saved++
	}

	p.log.Info("ingestion terminée", "recus", len(articles), "sauves", saved, "geolocalises", geocoded)
	return nil
}
