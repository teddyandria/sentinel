package geocoder

import (
	"context"
	"testing"
)

// TestStaticGeocoder_Geocode vérifie la détection de villes, l'insensibilité à
// la casse, la gestion des noms composés et l'absence de faux positifs.
func TestStaticGeocoder_Geocode(t *testing.T) {
	g := NewStaticGeocoder(DefaultCities())
	ctx := context.Background()

	tests := []struct {
		name     string
		text     string
		wantName string // "" => aucune localisation attendue
	}{
		{"ville simple", "Big tech event in London this week", "London"},
		{"ville composée", "Startup raises funds in San Francisco", "San Francisco"},
		{"insensible à la casse", "des rumeurs en provenance de paris", "Paris"},
		{"pas de faux positif", "a detailed comparison of web frameworks", ""},
		{"aucune ville", "an article without any known location", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := g.Geocode(ctx, tt.text)
			if err != nil {
				t.Fatalf("erreur inattendue: %v", err)
			}
			if tt.wantName == "" {
				if loc != nil {
					t.Fatalf("attendu aucune localisation, obtenu %+v", loc)
				}
				return
			}
			if loc == nil {
				t.Fatalf("attendu %q, obtenu nil", tt.wantName)
			}
			if loc.Name != tt.wantName {
				t.Fatalf("attendu %q, obtenu %q", tt.wantName, loc.Name)
			}
		})
	}
}
