// Package rag implémente la recherche "par le sens" + la rédaction d'une réponse :
//
//	question --embed--> vecteur --similarité--> top-k articles --LLM--> réponse + sources
//
// Il ne dépend que d'interfaces (Embedder, Generator, Store) : facile à tester
// et indépendant des implémentations concrètes (Ollama, Postgres).
package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/teddyandria/sentinel/internal/domain"
	"github.com/teddyandria/sentinel/internal/storage"
)

// Embedder transforme un texte en vecteur (le client Ollama le fait).
type Embedder interface {
	Embeddings(ctx context.Context, text string) ([]float32, error)
}

// Generator rédige du texte à partir d'un prompt (le client Ollama le fait).
type Generator interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Service câble l'embedder, le générateur et le stockage.
type Service struct {
	embedder  Embedder
	generator Generator
	store     storage.Store
}

func New(e Embedder, g Generator, s storage.Store) *Service {
	return &Service{embedder: e, generator: g, store: s}
}

// Answer = la réponse rédigée + les articles qui ont servi de sources.
type Answer struct {
	Answer  string           `json:"answer"`
	Sources []domain.Article `json:"sources"`
}

// Ask répond à une question : embed -> recherche des k plus proches -> rédaction.
func (svc *Service) Ask(ctx context.Context, question string, k int) (*Answer, error) {
	// 1. Vecteur de la question.
	qvec, err := svc.embedder.Embeddings(ctx, question)
	if err != nil {
		return nil, err
	}

	// 2. On charge les articles indexés et on garde les k plus proches du sens.
	docs, err := svc.store.ListEmbedded(ctx)
	if err != nil {
		return nil, err
	}
	sources := topK(qvec, docs, k)

	// 3. Le LLM rédige une réponse en s'appuyant UNIQUEMENT sur ces articles.
	text, err := svc.generator.Generate(ctx, buildPrompt(question, sources))
	if err != nil {
		return nil, err
	}

	return &Answer{Answer: strings.TrimSpace(text), Sources: sources}, nil
}

// topK renvoie les k articles dont le vecteur est le plus proche de qvec.
func topK(qvec []float32, docs []storage.EmbeddedArticle, k int) []domain.Article {
	type scored struct {
		article domain.Article
		score   float64
	}
	ranked := make([]scored, 0, len(docs))
	for _, d := range docs {
		ranked = append(ranked, scored{article: d.Article, score: cosine(qvec, d.Embedding)})
	}
	// Tri par similarité décroissante.
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	if k > len(ranked) {
		k = len(ranked)
	}
	out := make([]domain.Article, 0, k)
	for i := 0; i < k; i++ {
		out = append(out, ranked[i].article)
	}
	return out
}

// cosine mesure la proximité de sens entre deux vecteurs (1 = identique, 0 = sans rapport).
func cosine(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// buildPrompt assemble la consigne + les articles retenus pour le LLM.
func buildPrompt(question string, sources []domain.Article) string {
	var b strings.Builder
	b.WriteString("You are a news analyst. Answer the user's question using ONLY the articles below. ")
	b.WriteString("Be concise (3-4 sentences). If the articles don't answer the question, say so.\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nArticles:\n")
	for i, a := range sources {
		fmt.Fprintf(&b, "[%d] %s — %s\n", i+1, a.Title, a.Description)
	}
	b.WriteString("\nAnswer:")
	return b.String()
}
