// Package newsapi est un client HTTP pour l'API https://newsapi.org.
package newsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://newsapi.org/v2"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Article est la réponse brute de NewsAPI (mappée ensuite vers domain.Article).
type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Source      struct {
		Name string `json:"name"`
	} `json:"source"`
}

type everythingResponse struct {
	Status   string    `json:"status"` // "ok" ou "error"
	Message  string    `json:"message"`
	Articles []Article `json:"articles"`
}

// Everything interroge /everything et renvoie les articles d'une recherche.
func (c *Client) Everything(ctx context.Context, query string) ([]Article, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("sortBy", "publishedAt")
	params.Set("language", "en")
	params.Set("pageSize", strconv.Itoa(100)) // max autorisé par NewsAPI
	params.Set("apiKey", c.apiKey)

	endpoint := c.baseURL + "/everything?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("newsapi: création requête: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("newsapi: appel HTTP: %w", err)
	}
	defer resp.Body.Close()

	var out everythingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("newsapi: décodage réponse: %w", err)
	}

	// NewsAPI signale ses erreurs dans le corps (status="error") en plus du code HTTP.
	if resp.StatusCode != http.StatusOK || out.Status != "ok" {
		return nil, fmt.Errorf("newsapi: réponse en erreur (http %d): %s", resp.StatusCode, out.Message)
	}

	return out.Articles, nil
}
