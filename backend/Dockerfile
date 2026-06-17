# Image construite depuis la RACINE du repo (contexte de build = .) afin d'avoir
# accès à backend/ (code Go) et frontend/ (source React/Vite).
# Voir docker-compose.yml : context: . / dockerfile: backend/Dockerfile

# --- Étape 1 : build du frontend (Node + Vite) ---
FROM node:22-alpine AS web
WORKDIR /web
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build   # produit /web/dist

# --- Étape 2 : build du backend (Go) ---
FROM golang:1.26-alpine AS build
WORKDIR /src

# go.mod/go.sum d'abord pour profiter du cache Docker.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Binaire statique, exécutable dans une image minimale.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/sentinel ./cmd/sentinel

# --- Étape 3 : image finale ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /bin/sentinel /app/sentinel
COPY --from=web /web/dist /app/web

# Le binaire sert le frontend buildé depuis ce dossier.
ENV WEB_DIR=/app/web

EXPOSE 8080
ENTRYPOINT ["/app/sentinel"]
