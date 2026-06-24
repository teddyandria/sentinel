// Package mapbox est un client minimal pour l'API de géocodage Mapbox :
// transforme un nom de lieu en coordonnées (lon/lat).
package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.mapbox.com/geocoding/v5/mapbox.places"

type Geocoder struct {
	token      string
	httpClient *http.Client
}

func NewGeocoder(token string) *Geocoder {
	return &Geocoder{
		token:      token,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Place est le résultat d'un géocodage : un lieu canonique et ses coordonnées.
type Place struct {
	Name string  // libellé renvoyé par Mapbox (ex: "Lyon, France")
	Lat  float64
	Lon  float64
}

// geocodeResponse correspond au format JSON renvoyé par Mapbox.
// center est au format [longitude, latitude].
type geocodeResponse struct {
	Features []struct {
		PlaceName string    `json:"place_name"`
		Center    []float64 `json:"center"`
	} `json:"features"`
}

// Forward géocode un nom de lieu. Renvoie (nil, nil) si Mapbox ne trouve rien.
func (g *Geocoder) Forward(ctx context.Context, place string) (*Place, error) {
	// Le nom de lieu fait partie du chemin de l'URL : il doit être encodé.
	endpoint := fmt.Sprintf("%s/%s.json?limit=1&access_token=%s",
		baseURL, url.PathEscape(place), url.QueryEscape(g.token))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mapbox: appel HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mapbox: réponse en erreur (http %d)", resp.StatusCode)
	}

	var out geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("mapbox: décodage réponse: %w", err)
	}

	if len(out.Features) == 0 || len(out.Features[0].Center) < 2 {
		return nil, nil // aucun résultat
	}

	f := out.Features[0]
	return &Place{Name: f.PlaceName, Lon: f.Center[0], Lat: f.Center[1]}, nil
}
