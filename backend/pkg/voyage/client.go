// Package voyage est un client minimal pour l'API Voyage AI
// (https://voyageai.com), utilisé en production pour les embeddings —
// Anthropic n'a pas d'API d'embeddings native.
package voyage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func New(apiKey, model string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type embeddingsRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type embeddingsResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (c *Client) Embeddings(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(embeddingsRequest{Input: []string{text}, Model: c.model})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.voyageai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voyage: appel HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage: réponse en erreur (http %d)", resp.StatusCode)
	}

	var out embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("voyage: décodage réponse: %w", err)
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("voyage: réponse vide")
	}
	return out.Data[0].Embedding, nil
}
