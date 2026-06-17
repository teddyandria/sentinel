// Package domain contient les types métier partagés par les couches.
// Il ne dépend d'aucune autre couche pour éviter les dépendances circulaires.
package domain

import "time"

type Location struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type Article struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"` // lien vers la source, ouvert au clic sur le marqueur
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`

	Location *Location `json:"location,omitempty"` // nil tant que l'article n'est pas géocodé

	Hash string `json:"-"` // empreinte de déduplication, non exposée dans l'API
}
