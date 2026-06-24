package geocoder

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/teddyandria/sentinel/internal/domain"
	"github.com/teddyandria/sentinel/pkg/mapbox"
	"github.com/teddyandria/sentinel/pkg/ollama"
)

// LLMGeocoder géocode en 2 étapes : un petit LLM (Ollama) extrait le LIEU dont
// parle l'article, puis Mapbox traduit ce lieu en coordonnées exactes.
// Chaque outil fait ce qu'il sait faire : compréhension vs coordonnées fiables.
type LLMGeocoder struct {
	llm    *ollama.Client
	mapbox *mapbox.Geocoder
}

func NewLLMGeocoder(llm *ollama.Client, mb *mapbox.Geocoder) *LLMGeocoder {
	return &LLMGeocoder{llm: llm, mapbox: mb}
}

// Consigne donnée au LLM : extraire un seul lieu, ou null, en JSON strict.
const placePrompt = `You extract the single most relevant real-world location (city, region, or country) that the following news text is about.
Respond ONLY with JSON of the form {"place": "<place name>"} or {"place": null} if there is no clear location.

Text:
%s`

func (g *LLMGeocoder) Geocode(ctx context.Context, text string) (*domain.Location, error) {
	// 1. Le LLM renvoie un nom de lieu (ou null) au format JSON.
	raw, err := g.llm.GenerateJSON(ctx, strings.Replace(placePrompt, "%s", text, 1))
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Place string `json:"place"` // null -> "" automatiquement
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, nil // réponse non exploitable : on n'échoue pas tout le pipeline
	}
	place := strings.TrimSpace(parsed.Place)
	if place == "" {
		return nil, nil
	}

	// 2. Mapbox transforme le nom de lieu en coordonnées exactes.
	resolved, err := g.mapbox.Forward(ctx, place)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, nil // lieu introuvable côté Mapbox
	}

	return &domain.Location{Name: resolved.Name, Lat: resolved.Lat, Lon: resolved.Lon}, nil
}

// Vérifie à la compilation que *LLMGeocoder implémente Geocoder.
var _ Geocoder = (*LLMGeocoder)(nil)
