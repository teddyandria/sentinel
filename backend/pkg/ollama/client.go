// Package ollama est un client minimal pour un serveur Ollama local
// (https://ollama.com), utilisé pour faire tourner un petit LLM en local.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL     string
	model       string // modèle de géocodage (petit, fréquent : extraction de lieu)
	answerModel string // modèle de rédaction RAG (plus gros, occasionnel : réponses)
	embedModel  string // modèle d'embeddings (texte -> vecteur)
	httpClient  *http.Client
}

// New crée un client Ollama. Le timeout est généreux : le tout premier appel
// charge le modèle en mémoire (quelques secondes), les suivants sont rapides.
func New(baseURL, model, answerModel, embedModel string) *Client {
	return &Client{
		baseURL:     baseURL,
		model:       model,
		answerModel: answerModel,
		embedModel:  embedModel,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
	}
}

type generateRequest struct {
	Model     string         `json:"model"`
	Prompt    string         `json:"prompt"`
	Stream    bool           `json:"stream"`
	Format    string         `json:"format,omitempty"`     // "json" force une sortie JSON valide
	KeepAlive string         `json:"keep_alive,omitempty"` // décharge vite le modèle de la RAM
	Options   map[string]any `json:"options,omitempty"`
}

type generateResponse struct {
	Response string `json:"response"`
}

// GenerateJSON exige une réponse JSON valide (utilisé pour l'extraction de lieu).
// keep_alive court : le modèle est déchargé peu après, pour ne pas chauffer au repos.
func (c *Client) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, generateRequest{
		Model:     c.model,
		Prompt:    prompt,
		Stream:    false,
		Format:    "json",
		KeepAlive: "30s",
		Options:   map[string]any{"temperature": 0},
	})
}

// Generate produit du texte libre (utilisé pour rédiger la réponse RAG).
// Utilise le modèle de rédaction (plus gros) plutôt que celui du géocodage.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.generate(ctx, generateRequest{
		Model:     c.answerModel,
		Prompt:    prompt,
		Stream:    false,
		KeepAlive: "30s",
		Options:   map[string]any{"temperature": 0.2},
	})
}

// generate fait l'appel HTTP commun et renvoie le texte produit par le modèle.
func (c *Client) generate(ctx context.Context, reqBody generateRequest) (string, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: appel HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: réponse en erreur (http %d)", resp.StatusCode)
	}

	var out generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ollama: décodage réponse: %w", err)
	}
	return out.Response, nil
}

type embeddingsRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	KeepAlive string `json:"keep_alive,omitempty"`
}

type embeddingsResponse struct {
	Embedding []float32 `json:"embedding"`
}

// Embeddings transforme un texte en vecteur (liste de nombres) via le modèle
// d'embeddings. Deux textes au sens proche donnent deux vecteurs proches : c'est
// ce qui permet la recherche "par le sens" du RAG.
func (c *Client) Embeddings(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(embeddingsRequest{
		Model:     c.embedModel,
		Prompt:    text,
		KeepAlive: "30s",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: appel embeddings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: embeddings en erreur (http %d)", resp.StatusCode)
	}

	var out embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("ollama: décodage embeddings: %w", err)
	}
	return out.Embedding, nil
}
