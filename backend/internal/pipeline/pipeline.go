// Package pipeline orchestre l'ingestion : fetch -> geocode -> store.
package pipeline

import (
	"context"
	"log/slog"

	"github.com/teddyandria/sentinel/internal/fetcher"
	"github.com/teddyandria/sentinel/internal/geocoder"
	"github.com/teddyandria/sentinel/internal/storage"
)

// Embedder transforme un texte en vecteur (le client Ollama le fait).
type Embedder interface {
	Embeddings(ctx context.Context, text string) ([]float32, error)
}

type Pipeline struct {
	fetchers []fetcher.Fetcher
	geocoder geocoder.Geocoder
	embedder Embedder
	store    storage.Store
	log      *slog.Logger
}

func New(fetchers []fetcher.Fetcher, g geocoder.Geocoder, e Embedder, s storage.Store, log *slog.Logger) *Pipeline {
	return &Pipeline{fetchers: fetchers, geocoder: g, embedder: e, store: s, log: log}
}

// Run exécute une passe d'ingestion sur toutes les sources. Une source ou un
// article en erreur n'interrompt pas le reste du traitement.
func (p *Pipeline) Run(ctx context.Context) error {
	var received, skipped, saved, geocoded int

	for _, f := range p.fetchers {
		articles, err := f.Fetch(ctx)
		if err != nil {
			p.log.Error("fetch d'une source échoué", "err", err)
			continue
		}
		received += len(articles)

		for _, a := range articles {
			// On ne géocode QUE les nouveaux articles : un article déjà connu
			// ne repasse pas par le LLM (évite de faire chauffer la machine).
			known, err := p.store.Exists(ctx, a.Hash)
			if err != nil {
				p.log.Error("vérification existence échouée", "url", a.URL, "err", err)
				continue
			}
			if known {
				skipped++
				continue
			}

			loc, err := p.geocoder.Geocode(ctx, a.Title+" "+a.Description)
			if err != nil {
				p.log.Warn("géocodage échoué", "url", a.URL, "err", err)
			} else if loc != nil {
				a.Location = loc
				geocoded++
			}

			// Indexation : on calcule le vecteur de l'article pour la recherche RAG.
			if vec, err := p.embedder.Embeddings(ctx, a.Title+" "+a.Description); err != nil {
				p.log.Warn("embedding échoué", "url", a.URL, "err", err)
			} else {
				a.Embedding = vec
			}

			if err := p.store.Save(ctx, a); err != nil {
				p.log.Error("sauvegarde échouée", "url", a.URL, "err", err)
				continue
			}
			saved++
		}
	}

	p.log.Info("ingestion terminée", "recus", received, "ignores", skipped, "sauves", saved, "geolocalises", geocoded)
	return nil
}
