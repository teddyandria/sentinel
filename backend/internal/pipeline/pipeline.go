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
	fetchers []fetcher.Fetcher
	geocoder geocoder.Geocoder
	store    storage.Store
	log      *slog.Logger
}

func New(fetchers []fetcher.Fetcher, g geocoder.Geocoder, s storage.Store, log *slog.Logger) *Pipeline {
	return &Pipeline{fetchers: fetchers, geocoder: g, store: s, log: log}
}

// Run exécute une passe d'ingestion sur toutes les sources. Une source ou un
// article en erreur n'interrompt pas le reste du traitement.
func (p *Pipeline) Run(ctx context.Context) error {
	var received, saved, geocoded int

	for _, f := range p.fetchers {
		articles, err := f.Fetch(ctx)
		if err != nil {
			p.log.Error("fetch d'une source échoué", "err", err)
			continue
		}
		received += len(articles)

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
	}

	p.log.Info("ingestion terminée", "recus", received, "sauves", saved, "geolocalises", geocoded)
	return nil
}
