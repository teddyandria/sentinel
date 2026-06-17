# 🛰️ Sentinel

**Sentinel** est un outil de **veille tech/actualité automatisée** qui récupère des
articles depuis plusieurs sources, en extrait la **localisation géographique**, puis
les affiche sur une **carte interactive cliquable** : chaque point = un article,
un clic = redirection vers la source.

> Projet personnel, en Go, pensé pour itérer vite tout en gardant une séparation
> claire des responsabilités.

---

## ✨ Fonctionnalités (vision)

1. **Fetcher** — récupère des articles depuis plusieurs sources (NewsAPI ; RSS / scraping à venir)
2. **Enrichissement** — extraction de la localisation mentionnée (règles simples, puis NER)
3. **Stockage** — persistance en PostgreSQL avec **déduplication**
4. **API HTTP** — expose les articles géolocalisés en JSON
5. **Frontend** — carte interactive Mapbox avec marqueurs cliquables
6. **Scheduler** — exécution périodique du fetch (cron interne)

> ℹ️ L'état actuel du dépôt est un **squelette structurel** : l'architecture, les
> interfaces et le câblage sont en place ; la logique métier (`TODO(impl)`) reste à écrire.

---

## 🧱 Stack & choix techniques

| Domaine | Choix | Justification |
|---|---|---|
| Langage | **Go** | Performant, simple à déployer (binaire unique), concurrence native |
| Routeur HTTP | **chi/v5** | Léger, 100 % compatible `net/http`, middlewares propres |
| Logger | **slog** (stdlib) | Structuré, zéro dépendance |
| Config | **env vars + godotenv** | 12-factor, explicite, léger (vs Viper) |
| Base de données | **PostgreSQL** via **pgx/v5** | Driver natif performant, API moderne |
| Scheduler | **`time.Ticker`** (stdlib) | Suffisant pour une période fixe |
| Frontend | **HTML/JS vanilla + Mapbox GL JS** | Pas de build step, carte vectorielle WebGL fluide |
| Conteneurisation | **Docker + docker-compose** | Environnement reproductible |

---

## 📁 Structure du projet

Le **backend** et le **frontend** sont dans deux dossiers distincts. En dev comme
en conteneur, le serveur Go sert aussi les fichiers statiques du front.

```
sentinel/
├── backend/                    # Tout le code Go (module à part entière)
│   ├── cmd/
│   │   └── sentinel/
│   │       └── main.go         # Entrée : config, logger, câblage, graceful shutdown
│   ├── internal/               # Code privé à l'application
│   │   ├── api/                # Serveur HTTP (chi) + handlers JSON
│   │   ├── config/             # Chargement de la configuration (env)
│   │   ├── domain/             # Types métier (Article, Location) — sans dépendances
│   │   ├── fetcher/            # Interface Fetcher + implémentation NewsAPI
│   │   ├── geocoder/           # Interface Geocoder + géocodeur statique
│   │   ├── scheduler/          # Exécution périodique du pipeline
│   │   └── storage/            # Interface Store + implémentation Postgres (pgx)
│   ├── pkg/
│   │   └── newsapi/            # Client NewsAPI réutilisable (importable ailleurs)
│   ├── migrations/             # Schéma SQL (joué au 1er démarrage de Postgres)
│   ├── Dockerfile              # Build multi-stage, image finale légère
│   └── go.mod
├── frontend/                   # Carte interactive (statique)
│   ├── index.html
│   ├── app.js                  # Charge /api/config + /api/articles, marqueurs Mapbox
│   └── style.css
├── docker-compose.yml          # Postgres + app
├── Makefile                    # run, build, test, lint, docker-up...
├── .env.example                # Modèle de configuration
└── README.md
```

### Pourquoi `internal/` et `pkg/` ?

- **`internal/`** : code privé au module ; Go interdit son import depuis l'extérieur.
  C'est là que vit la logique propre à Sentinel.
- **`pkg/`** : code réutilisable et autonome. Le client `newsapi` n'a aucune dépendance
  vers Sentinel : il pourrait être extrait dans un autre projet.
- **`domain/`** ne dépend d'aucune autre couche : c'est le cœur du modèle vers lequel
  tout pointe (évite les dépendances circulaires).

Chaque couche clé (`Fetcher`, `Store`, `Geocoder`) est définie par une **interface**,
ce qui permet de la mocker en test et d'en changer l'implémentation sans toucher au reste.

---

## 🚀 Démarrage rapide

### Prérequis
- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) + Docker Compose
- Une clé [NewsAPI](https://newsapi.org) (gratuite)
- Un token public [Mapbox](https://account.mapbox.com) (`pk.*`, gratuit) pour la carte

### 1. Configuration
```bash
cp .env.example .env
# puis renseigne dans .env :
#   NEWS_API_KEY  (source des articles)
#   MAPBOX_TOKEN  (token public pk.* pour afficher la carte)
```

### 2. Tout lancer avec Docker (recommandé)
```bash
make docker-up      # démarre Postgres + l'application
# Carte : http://localhost:8080
# API   : http://localhost:8080/api/articles
make docker-down    # arrêt
```

### 3. Lancer en local (sans conteneuriser l'app)
Il faut un Postgres accessible (ex: `make docker-up` ne lançant que `db`, ou un Postgres local)
puis :
```bash
make run            # exécute le backend, sert le front depuis ../frontend
```

---

## 🔌 API

| Méthode | Route | Description |
|---|---|---|
| `GET` | `/api/health` | Sonde de vivacité (`{"status":"ok"}`) |
| `GET` | `/api/config` | Config publique du front (`{"mapboxToken":"..."}`) |
| `GET` | `/api/articles` | Liste des articles géolocalisés (JSON) |
| `GET` | `/*` | Frontend statique (carte Mapbox) |

Exemple de réponse `/api/articles` :
```json
[
  {
    "id": 1,
    "title": "Exemple d'article",
    "description": "...",
    "url": "https://source.example/article",
    "source": "TechNews",
    "published_at": "2026-06-05T10:00:00Z",
    "location": { "name": "Paris", "lat": 48.8566, "lon": 2.3522 }
  }
]
```

---

## 🛠️ Commandes Make

```bash
make            # affiche l'aide
make run        # lance l'app en local
make build      # compile le binaire (backend/bin/)
make test       # tests + race detector + couverture
make lint       # go vet (+ golangci-lint si installé)
make tidy       # met à jour go.mod / go.sum
make docker-up  # Postgres + app via docker-compose
make docker-down
make clean
```

---

## ✅ Implémenté

- [x] `newsapi.Client.Everything` (appel HTTP réel à NewsAPI)
- [x] `NewsAPIFetcher.Fetch` : mapping vers `domain.Article` + `Hash` (dédup)
- [x] `StaticGeocoder.Geocode` : dictionnaire de villes, matching par limites de mots
- [x] `PostgresStore.Save` (INSERT ... ON CONFLICT) et `ListGeolocated`
- [x] `pipeline.Run` : fetch → geocode → store, branché au scheduler
- [x] Frontend Mapbox (token servi par `/api/config`)

## 🗺️ Prochaines étapes

- [ ] Enrichir le dictionnaire de villes (ou passer à un vrai NER / une API de géocodage)
- [ ] Ajouter la pagination NewsAPI et d'autres sources (RSS, scraping)
- [ ] Tests d'intégration storage (Postgres via testcontainers)
- [ ] CI/CD GitHub Actions (build, test, lint)
