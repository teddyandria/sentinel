// Command reindex calcule le vecteur des articles qui n'en ont pas encore
// (rattrapage de l'historique pour la recherche RAG). Traité par petits lots
// avec une courte pause, pour ne pas faire chauffer la machine.
package main

import (
	"context"
	"log"
	"time"

	"github.com/teddyandria/sentinel/internal/config"
	"github.com/teddyandria/sentinel/internal/storage"
	"github.com/teddyandria/sentinel/pkg/ollama"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	store, err := storage.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	llm := ollama.New(cfg.OllamaURL, cfg.OllamaModel, cfg.OllamaAnswer, cfg.OllamaEmbed)

	const batchSize = 50
	total := 0

	for {
		articles, err := store.ArticlesToEmbed(ctx, batchSize)
		if err != nil {
			log.Fatal(err)
		}
		if len(articles) == 0 {
			break // plus rien à indexer
		}

		for _, a := range articles {
			vec, err := llm.Embeddings(ctx, a.Title+" "+a.Description)
			if err != nil {
				log.Printf("embedding échoué (id=%d): %v", a.ID, err)
				continue
			}
			if err := store.SaveEmbedding(ctx, a.ID, vec); err != nil {
				log.Printf("sauvegarde échouée (id=%d): %v", a.ID, err)
				continue
			}
			total++
		}
		log.Printf("indexés: %d", total)
		time.Sleep(200 * time.Millisecond) // petite pause anti-chauffe entre les lots
	}

	log.Printf("ré-indexation terminée : %d articles indexés", total)
}
