# Image construite depuis la RACINE du repo (contexte de build = .) afin d'avoir
# accès à la fois à backend/ (code Go) et frontend/ (fichiers statiques).
# Voir docker-compose.yml : context: . / dockerfile: backend/Dockerfile

# --- Étape 1 : build ---
# Image Go complète, uniquement pour compiler. Elle ne se retrouve pas dans l'image finale.
FROM golang:1.26-alpine AS build

WORKDIR /src

# On copie d'abord go.mod/go.sum pour profiter du cache Docker :
# tant que les dépendances ne changent pas, cette couche est réutilisée.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Puis le reste du code backend.
COPY backend/ ./

# CGO_ENABLED=0 -> binaire statique, exécutable dans une image minimale.
# -ldflags "-s -w" retire les infos de debug pour alléger le binaire.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/sentinel ./cmd/sentinel

# --- Étape 2 : image finale ---
# Image légère : seulement le binaire, les certificats CA (HTTPS sortant vers NewsAPI)
# et le frontend statique.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /bin/sentinel /app/sentinel
COPY frontend /app/frontend

# Le binaire sert les fichiers statiques depuis ce dossier (voir WEB_DIR).
ENV WEB_DIR=/app/frontend

EXPOSE 8080
ENTRYPOINT ["/app/sentinel"]
