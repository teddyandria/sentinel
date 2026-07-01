// Package pipeline orchestre l'ingestion : fetch -> geocode -> store.
package pipeline

import (
	"context"
	"log/slog"
	"time"

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
	store    storage.Store
	log      *slog.Logger
}

func New(fetchers []fetcher.Fetcher, g geocoder.Geocoder, s storage.Store, log *slog.Logger) *Pipeline {
	return &Pipeline{fetchers: fetchers, geocoder: g, store: s, log: log}
}

// Indexer calcule et stocke les embeddings manquants en différé, par petits
// lots espacés, pour rester sous le rate limit de l'API d'embeddings.
type Indexer struct {
	embedder  Embedder
	store     storage.Store
	log       *slog.Logger
	batchSize int
	pause     time.Duration
}

func NewIndexer(e Embedder, s storage.Store, log *slog.Logger) *Indexer {
	return &Indexer{embedder: e, store: s, log: log, batchSize: 10, pause: time.Second}
}

func (ix *Indexer) Run(ctx context.Context) error {
	articles, err := ix.store.ArticlesToEmbed(ctx, ix.batchSize)
	if err != nil {
		return err
	}
	for _, a := range articles {
		vec, err := ix.embedder.Embeddings(ctx, a.Title+" "+a.Description)
		if err != nil {
			ix.log.Warn("embedding différé échoué", "id", a.ID, "err", err)
			continue
		}
		if err := ix.store.SaveEmbedding(ctx, a.ID, vec); err != nil {
			ix.log.Warn("sauvegarde embedding échouée", "id", a.ID, "err", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(ix.pause):
		}
	}
	if len(articles) > 0 {
		ix.log.Info("indexation différée", "indexés", len(articles))
	}
	return nil
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

			if err := p.store.Save(ctx, a); err != nil {
				p.log.Error("sauvegarde échouée", "url", a.URL, "err", err)
				continue
			}
			saved++
			// Pause entre chaque article pour respecter le rate limit de Groq (géocodage).
			time.Sleep(200 * time.Millisecond)
		}
	}

	p.log.Info("ingestion terminée", "recus", received, "ignores", skipped, "sauves", saved, "geolocalises", geocoded)
	return nil
}
