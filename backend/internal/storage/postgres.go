package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/teddyandria/sentinel/internal/domain"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	// Ping pour échouer tôt si la base est injoignable.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

// Save insère un article ; ON CONFLICT (hash) DO NOTHING ignore les doublons.
func (s *PostgresStore) Save(ctx context.Context, a domain.Article) error {
	// Location (pointeur, possiblement nil) -> colonnes nullables.
	var locName *string
	var lat, lon *float64
	if a.Location != nil {
		locName = &a.Location.Name
		lat = &a.Location.Lat
		lon = &a.Location.Lon
	}

	const q = `
		INSERT INTO articles (title, description, url, image_url, source, topic, published_at, location_name, lat, lon, hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (hash) DO NOTHING`

	_, err := s.pool.Exec(ctx, q,
		a.Title, a.Description, a.URL, a.ImageURL, a.Source, a.Topic, a.PublishedAt, locName, lat, lon, a.Hash,
	)
	return err
}

func (s *PostgresStore) ListGeolocated(ctx context.Context, topic string) ([]domain.Article, error) {
	// Requête construite dynamiquement : le filtre topic est optionnel.
	q := `
		SELECT id, title, description, url, image_url, source, topic, published_at, location_name, lat, lon
		FROM articles
		WHERE lat IS NOT NULL AND lon IS NOT NULL`
	var args []any
	if topic != "" {
		q += " AND topic = $1"
		args = append(args, topic)
	}
	q += " ORDER BY published_at DESC LIMIT 500"

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		// Colonnes nullables scannées dans des pointeurs.
		var description, imageURL, source, locName *string
		var lat, lon *float64

		if err := rows.Scan(
			&a.ID, &a.Title, &description, &a.URL, &imageURL, &source, &a.Topic, &a.PublishedAt, &locName, &lat, &lon,
		); err != nil {
			return nil, err
		}

		if description != nil {
			a.Description = *description
		}
		if imageURL != nil {
			a.ImageURL = *imageURL
		}
		if source != nil {
			a.Source = *source
		}
		if lat != nil && lon != nil {
			loc := domain.Location{Lat: *lat, Lon: *lon}
			if locName != nil {
				loc.Name = *locName
			}
			a.Location = &loc
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

// Vérifie à la compilation que *PostgresStore implémente Store.
var _ Store = (*PostgresStore)(nil)
