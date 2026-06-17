// Package fetcher récupère des articles depuis des sources externes (NewsAPI, RSS...).
package fetcher

import (
	"context"

	"github.com/teddyandria/sentinel/internal/domain"
)

// Fetcher abstrait une source d'articles : chaque source concrète l'implémente,
// ce qui permet de la remplacer par un mock dans les tests.
type Fetcher interface {
	Fetch(ctx context.Context) ([]domain.Article, error)
}
