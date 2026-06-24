// Package storage persiste les articles dans Postgres avec déduplication.
package storage

import (
	"context"

	"github.com/teddyandria/sentinel/internal/domain"
)

// Store abstrait la persistance : permet de remplacer Postgres par un mock en test.
type Store interface {
	// Save insère un article ; la déduplication se fait via Article.Hash.
	Save(ctx context.Context, a domain.Article) error
	// Exists indique si un article (identifié par son Hash) est déjà en base.
	// Sert à ne pas re-géocoder un article déjà connu (coûteux en LLM).
	Exists(ctx context.Context, hash string) (bool, error)
	// ListGeolocated renvoie les articles géolocalisés ; topic="" = tous les sujets.
	ListGeolocated(ctx context.Context, topic string) ([]domain.Article, error)
	// ListEmbedded charge les articles indexés (article + vecteur) pour la recherche RAG.
	ListEmbedded(ctx context.Context) ([]EmbeddedArticle, error)
	Close() error
}
