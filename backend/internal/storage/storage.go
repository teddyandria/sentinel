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
	ListGeolocated(ctx context.Context) ([]domain.Article, error)
	Close() error
}
