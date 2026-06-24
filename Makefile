# Makefile Sentinel — raccourcis de développement.
# Utilisation : make <cible>   (ex: make run). « make » seul affiche l'aide.
#
# Le code Go vit dans backend/ : les cibles Go s'exécutent donc dans ce dossier.

# Charge le .env pour les cibles locales (run). Optionnel grâce au "-".
-include .env
export

BACKEND  := backend
FRONTEND := frontend
BINARY   := bin/sentinel

.PHONY: help run db-up reindex front front-install front-build build test lint tidy docker-up docker-down clean

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: db-up ## Lance le backend (démarre d'abord la base si besoin)
	cd $(BACKEND) && go run ./cmd/sentinel

reindex: db-up ## Calcule les vecteurs des articles sans embedding (recherche RAG)
	cd $(BACKEND) && go run ./cmd/reindex

db-up: ## Démarre Postgres (et Docker Desktop si éteint), attend qu'il soit prêt
	@docker info >/dev/null 2>&1 || { echo "Démarrage de Docker Desktop..."; \
		open -a Docker 2>/dev/null || open -a "Docker Desktop" 2>/dev/null; \
		for i in $$(seq 1 45); do docker info >/dev/null 2>&1 && break; sleep 3; done; }
	@docker compose up -d db
	@printf "Attente de Postgres"; \
	for i in $$(seq 1 20); do \
		[ "$$(docker inspect -f '{{.State.Health.Status}}' sentinel2-db-1 2>/dev/null)" = "healthy" ] && { echo " OK"; break; }; \
		printf "."; sleep 2; \
	done

front-install: ## Installe les dépendances npm du frontend
	cd $(FRONTEND) && npm install

front: ## Lance le frontend en dev (Vite, http://localhost:5173, proxy /api -> :8080)
	cd $(FRONTEND) && npm run dev

front-build: ## Build le frontend dans frontend/dist
	cd $(FRONTEND) && npm run build

build: ## Compile le binaire dans backend/bin/
	cd $(BACKEND) && go build -o $(BINARY) ./cmd/sentinel

test: ## Lance les tests (avec détecteur de data race et couverture)
	cd $(BACKEND) && go test ./... -race -cover

lint: ## Analyse statique (go vet, + golangci-lint s'il est installé)
	cd $(BACKEND) && go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && (cd $(BACKEND) && golangci-lint run) || echo "golangci-lint non installé (optionnel)"

tidy: ## Met à jour go.mod / go.sum
	cd $(BACKEND) && go mod tidy

docker-up: ## Démarre Postgres + app via docker-compose
	docker compose up --build -d

docker-down: ## Arrête et supprime les conteneurs
	docker compose down

clean: ## Supprime les artefacts de build (binaire + dist front)
	rm -rf $(BACKEND)/bin/ $(FRONTEND)/dist/

# Cible par défaut : l'aide.
.DEFAULT_GOAL := help
