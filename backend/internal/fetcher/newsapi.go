package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"github.com/teddyandria/sentinel/internal/domain"
	"github.com/teddyandria/sentinel/pkg/newsapi"
)

// NewsAPIFetcher est une source = un fournisseur (NewsAPI) + un topic. Chaque
// instance ne récupère que les articles de son topic et les tague en conséquence.
type NewsAPIFetcher struct {
	client *newsapi.Client
	topic  string
	log    *slog.Logger
}

func NewNewsAPIFetcher(client *newsapi.Client, topic string, log *slog.Logger) *NewsAPIFetcher {
	return &NewsAPIFetcher{client: client, topic: topic, log: log}
}

// Fetch interroge NewsAPI pour le topic, ignore les articles inexploitables
// (sans titre/URL), les tague avec le topic et calcule un Hash pour la déduplication.
func (f *NewsAPIFetcher) Fetch(ctx context.Context) ([]domain.Article, error) {
	f.log.Debug("fetch NewsAPI", "topic", f.topic)

	raw, err := f.client.Everything(ctx, f.topic)
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
			ImageURL:    r.ImageURL,
			Source:      r.Source.Name,
			Topic:       f.topic,
			PublishedAt: r.PublishedAt,
			Hash:        hashURL(r.URL),
		})
	}

	f.log.Info("articles récupérés", "source", "newsapi", "topic", f.topic, "count", len(articles))
	return articles, nil
}

// hashURL calcule une empreinte stable d'une URL, utilisée comme clé de déduplication.
func hashURL(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

// Vérifie à la compilation que *NewsAPIFetcher implémente Fetcher.
var _ Fetcher = (*NewsAPIFetcher)(nil)
