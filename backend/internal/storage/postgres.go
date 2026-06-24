package storage

import (
	"context"
	"encoding/json"

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

	// Le vecteur (s'il existe) est stocké en JSON ; sinon NULL.
	var embedding *string
	if len(a.Embedding) > 0 {
		if b, err := json.Marshal(a.Embedding); err == nil {
			s := string(b)
			embedding = &s
		}
	}

	const q = `
		INSERT INTO articles (title, description, url, image_url, source, topic, published_at, location_name, lat, lon, hash, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (hash) DO NOTHING`

	_, err := s.pool.Exec(ctx, q,
		a.Title, a.Description, a.URL, a.ImageURL, a.Source, a.Topic, a.PublishedAt, locName, lat, lon, a.Hash, embedding,
	)
	return err
}

// EmbeddedArticle regroupe un article et son vecteur de sens, pour la recherche.
type EmbeddedArticle struct {
	Article   domain.Article
	Embedding []float32
}

// SaveEmbedding enregistre (après coup) le vecteur d'un article déjà en base.
// Utilisé par la commande de ré-indexation pour rattraper l'historique.
func (s *PostgresStore) SaveEmbedding(ctx context.Context, id int64, vec []float32) error {
	b, err := json.Marshal(vec)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `UPDATE articles SET embedding = $1 WHERE id = $2`, string(b), id)
	return err
}

// ArticlesToEmbed renvoie des articles encore sans vecteur (à indexer), par lots.
func (s *PostgresStore) ArticlesToEmbed(ctx context.Context, limit int) ([]domain.Article, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, title, description FROM articles WHERE embedding IS NULL ORDER BY id LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Article
	for rows.Next() {
		var a domain.Article
		var description *string
		if err := rows.Scan(&a.ID, &a.Title, &description); err != nil {
			return nil, err
		}
		if description != nil {
			a.Description = *description
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListEmbedded charge tous les articles déjà indexés (article + vecteur), pour
// que le retriever calcule la similarité avec la question.
func (s *PostgresStore) ListEmbedded(ctx context.Context) ([]EmbeddedArticle, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, description, url, image_url, source, topic, published_at, location_name, lat, lon, embedding
		FROM articles
		WHERE embedding IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []EmbeddedArticle
	for rows.Next() {
		var a domain.Article
		var description, imageURL, source, locName, embedding *string
		var lat, lon *float64

		if err := rows.Scan(
			&a.ID, &a.Title, &description, &a.URL, &imageURL, &source, &a.Topic, &a.PublishedAt, &locName, &lat, &lon, &embedding,
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

		var vec []float32
		if embedding != nil {
			_ = json.Unmarshal([]byte(*embedding), &vec)
		}
		out = append(out, EmbeddedArticle{Article: a, Embedding: vec})
	}
	return out, rows.Err()
}

// Exists indique si un article avec ce hash est déjà en base.
func (s *PostgresStore) Exists(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM articles WHERE hash = $1)`, hash).Scan(&exists)
	return exists, err
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
