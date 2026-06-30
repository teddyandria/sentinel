// Package groq est un client minimal pour l'API Groq (https://groq.com),
// utilisé en production pour la génération — gratuit (rate-limit généreux,
// sans carte bancaire), fait tourner des modèles open-source très vite.
package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiKey      string
	model       string // modèle de géocodage (rapide/économique)
	answerModel string // modèle de rédaction RAG (qualité de rédaction)
	httpClient  *http.Client
}

func New(apiKey, model, answerModel string) *Client {
	return &Client{
		apiKey:      apiKey,
		model:       model,
		answerModel: answerModel,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

// GenerateJSON force une sortie JSON valide (extraction de lieu) via le mode
// natif "json_object" de l'API Groq (compatible OpenAI).
func (c *Client) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	return c.chat(ctx, c.model, prompt, 0, map[string]any{"type": "json_object"})
}

// Generate produit du texte libre (réponse RAG), avec le modèle de rédaction.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.chat(ctx, c.answerModel, prompt, 0.2, nil)
}

func (c *Client) chat(ctx context.Context, model, prompt string, temperature float64, responseFormat map[string]any) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:          model,
		Messages:       []message{{Role: "user", Content: prompt}},
		Temperature:    temperature,
		ResponseFormat: responseFormat,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq: appel HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq: réponse en erreur (http %d)", resp.StatusCode)
	}

	var out chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("groq: décodage réponse: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("groq: réponse vide")
	}
	return out.Choices[0].Message.Content, nil
}
