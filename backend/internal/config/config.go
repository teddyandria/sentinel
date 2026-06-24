// Package config charge la configuration de l'application depuis l'environnement.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr      string        // adresse d'écoute du serveur HTTP (ex: ":8080")
	DatabaseURL   string        // DSN Postgres (postgres://user:pass@host:5432/db)
	NewsAPIKey    string        // clé d'API NewsAPI
	FetchInterval time.Duration // période entre deux exécutions du scheduler
	LogLevel      string        // niveau de log : debug, info, warn, error
	WebDir        string        // dossier du frontend statique servi par l'API
	MapboxToken   string        // token public Mapbox (pk.*) transmis au frontend
	OllamaURL     string        // URL du serveur Ollama local
	OllamaModel   string        // modèle de géocodage (petit, fréquent)
	OllamaAnswer  string        // modèle de rédaction RAG (plus gros, occasionnel)
	OllamaEmbed   string        // modèle d'embeddings (texte -> vecteur, pour le RAG)
}

func Load() (Config, error) {
	// godotenv est best-effort : en prod les variables viennent du système, en dev du .env.
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		NewsAPIKey:    getEnv("NEWS_API_KEY", ""),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		FetchInterval: getEnvDuration("FETCH_INTERVAL", 30*time.Minute),
		// Front buildé par Vite. En conteneur, on surcharge via WEB_DIR (ex: /app/web).
		WebDir:      getEnv("WEB_DIR", "../frontend/dist"),
		MapboxToken: getEnv("MAPBOX_TOKEN", ""),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:  getEnv("OLLAMA_MODEL", "llama3.2:1b"),
		OllamaAnswer: getEnv("OLLAMA_ANSWER_MODEL", "llama3.2:3b"),
		OllamaEmbed:  getEnv("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
	}

	if cfg.NewsAPIKey == "" {
		return Config{}, fmt.Errorf("config: NEWS_API_KEY est obligatoire")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL est obligatoire")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
