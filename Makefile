# Makefile Sentinel — raccourcis de développement.
# Utilisation : make <cible>   (ex: make run). « make » seul affiche l'aide.
#
# Le code Go vit dans backend/ : les cibles Go s'exécutent donc dans ce dossier.

# Charge le .env pour les cibles locales (run). Optionnel grâce au "-".
-include .env
export

BACKEND := backend
BINARY  := bin/sentinel

.PHONY: help run build test lint tidy docker-up docker-down clean

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

run: ## Lance l'application en local (sert aussi le front depuis ../frontend)
	cd $(BACKEND) && go run ./cmd/sentinel

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

clean: ## Supprime les artefacts de build
	rm -rf $(BACKEND)/bin/

# Cible par défaut : l'aide.
.DEFAULT_GOAL := help
