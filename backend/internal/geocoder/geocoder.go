// Package geocoder extrait une localisation à partir d'un texte libre.
package geocoder

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/teddyandria/sentinel/internal/domain"
)

type Geocoder interface {
	// Geocode renvoie la localisation détectée, ou (nil, nil) si rien n'est trouvé.
	Geocode(ctx context.Context, text string) (*domain.Location, error)
}

// StaticGeocoder détecte les villes d'une table de correspondance "nom -> coordonnées".
type StaticGeocoder struct {
	table map[string]domain.Location
	re    *regexp.Regexp
}

// NewStaticGeocoder pré-compile une regex qui repère n'importe quel nom de la table,
// bornée par des limites de mots \b pour éviter les faux positifs ("paris" dans
// "comparison"). Les noms longs passent en premier pour primer ("New York" > "York").
func NewStaticGeocoder(table map[string]domain.Location) *StaticGeocoder {
	normalized := make(map[string]domain.Location, len(table))
	names := make([]string, 0, len(table))
	for name, loc := range table {
		key := strings.ToLower(name)
		normalized[key] = loc
		names = append(names, key)
	}

	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })

	var re *regexp.Regexp
	if len(names) > 0 {
		quoted := make([]string, len(names))
		for i, n := range names {
			quoted[i] = regexp.QuoteMeta(n)
		}
		re = regexp.MustCompile(`\b(` + strings.Join(quoted, "|") + `)\b`)
	}

	return &StaticGeocoder{table: normalized, re: re}
}

func (g *StaticGeocoder) Geocode(ctx context.Context, text string) (*domain.Location, error) {
	if g.re == nil {
		return nil, nil
	}
	match := g.re.FindString(strings.ToLower(text))
	if match == "" {
		return nil, nil
	}
	loc, ok := g.table[match]
	if !ok {
		return nil, nil
	}
	out := loc // copie pour que l'appelant ne puisse pas modifier la table
	return &out, nil
}

// Vérifie à la compilation que *StaticGeocoder implémente Geocoder.
var _ Geocoder = (*StaticGeocoder)(nil)
