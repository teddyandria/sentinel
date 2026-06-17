package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"github.com/teddyandria/sentinel/internal/domain"
	"github.com/teddyandria/sentinel/pkg/newsapi"
)

type NewsAPIFetcher struct {
	client *newsapi.Client
	query  string
	log    *slog.Logger
}

func NewNewsAPIFetcher(client *newsapi.Client, query string, log *slog.Logger) *NewsAPIFetcher {
	return &NewsAPIFetcher{client: client, query: query, log: log}
}

// Fetch interroge NewsAPI, ignore les articles inexploitables (sans titre/URL)
// et calcule un Hash par article pour la déduplication.
func (f *NewsAPIFetcher) Fetch(ctx context.Context) ([]domain.Article, error) {
	f.log.Debug("fetch NewsAPI", "query", f.query)

	raw, err := f.client.Everything(ctx, f.query)
	if err != nil {
		return nil, err
	}

	articles := make([]domain.Article, 0, len(raw))
	for _, r := range raw {
		if r.Title == "" || r.URL == "" {
			continue
		}
		articles = append(articles, domain.Article{
			Title:       r.Title,
			Description: r.Description,
			URL:         r.URL,
			Source:      r.Source.Name,
			PublishedAt: r.PublishedAt,
			Hash:        hashURL(r.URL),
		})
	}

	f.log.Info("articles récupérés", "source", "newsapi", "count", len(articles))
	return articles, nil
}

// hashURL calcule une empreinte stable d'une URL, utilisée comme clé de déduplication.
func hashURL(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

// Vérifie à la compilation que *NewsAPIFetcher implémente Fetcher.
var _ Fetcher = (*NewsAPIFetcher)(nil)
