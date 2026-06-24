# Sentinel — Comment ça marche (expliqué simplement)

Ce document explique chaque brique de Sentinel : **le problème qu'elle règle**, et **ce qu'elle fait**.
Pas de jargon, des images.

---

## Vue d'ensemble

Sentinel récupère des actualités, devine **de quel lieu** chacune parle, et les affiche sur une **carte cliquable**.

```
NewsAPI ──► Fetcher ──► Géocodage ──► Postgres ──► API ──► Carte
(actus)    (récupère)  (trouve le    (range +     (sert    (affiche les
                        lieu)         dédoublonne)  en JSON) points)
                         ▲
                         │ déclenché toutes les 30 min par le Scheduler
```

Le fil rouge de l'architecture : **chaque brique est remplaçable** sans toucher aux autres
(grâce aux *interfaces* Go). On l'a prouvé en remplaçant le géocodage « dictionnaire » par un géocodage « IA » en changeant **une seule ligne**.

---

## 1. Le Fetcher — récupérer les actus

**Le problème.** Les actualités sont éparpillées sur des sources externes, et on veut plusieurs sujets (tech, business, science, santé, politique).

**Ce qu'on fait.** Une *source* = un fournisseur (NewsAPI) + un sujet. On crée un fetcher par sujet ; chacun récupère ses articles et les **étiquette** avec son sujet.

```
NewsAPIFetcher{technology} ─┐
NewsAPIFetcher{business}   ─┤
NewsAPIFetcher{science}    ─┼─► le pipeline les interroge tous
NewsAPIFetcher{health}     ─┤
NewsAPIFetcher{politics}   ─┘
```

> Demain, un fetcher RSS étiqueté « science » entrera dans la même liste sans rien casser.

---

## 2. La déduplication — pas deux fois le même article

**Le problème.** La même actu revient à chaque récupération (et parfois republiée par plusieurs médias). On ne veut pas 10 fois le même point.

**Ce qu'on fait.** Pour chaque article on calcule une **empreinte** (un hash de son URL). En base, cette empreinte est **unique** : un article déjà connu est simplement ignoré.

```
URL de l'article ──► empreinte "a3f9c2…" ──► déjà en base ? ──► oui : on ignore / non : on enregistre
```

---

## 3. Le Géocodage — deviner le lieu (la pièce maîtresse)

### Le problème

Une carte a besoin de savoir **où** poser chaque point. Mais NewsAPI ne donne **pas** de coordonnées — juste du texte.

Notre première version était un **videur de boîte de nuit avec une liste de 15 noms** :

- « Paris » → sur la liste → **entre** ✅
- « la capitale française » → pas écrit pareil → **refusé** ❌
- « Lyon » → pas dans les 15 → **refusé** ❌
- « Silicon Valley » → **refusé** ❌

Il **comparait des mots, sans comprendre**. Résultat : **~94 % des articles restaient invisibles** sur la carte.

### Ce qu'on fait : un LLM + Mapbox

On remplace le videur à la liste par **un humain malin qui a fait de la géo**. On lui donne l'article et on demande : **« ça parle d'où ? »**. Il comprend, même si c'est tourné autrement :

| L'article dit… | Videur à la liste | LLM (humain malin) |
|---|---|---|
| « …a store in **Tokyo** » | « Tokyo » ✅ | « Tokyo » ✅ |
| « …in the **French capital** » | rien ❌ | « Paris » ✅ |
| « …based in **Lyon** » | rien ❌ | « Lyon » ✅ |
| « …**Silicon Valley** startups » | rien ❌ | « San Francisco » ✅ |
| « …**Parisian** fashion week » | rien ❌ | « Paris » ✅ |

Mais le LLM ne donne que le **nom** (« Lyon »). Une carte a besoin de **chiffres**. C'est le rôle de **Mapbox** (un GPS) : « Lyon » → `45.76, 4.83`.

```
Article ─► 🧠 LLM : « ça parle de Lyon » ─► 📍 Mapbox : 45.76, 4.83 ─► point sur la carte
          (comprend le lieu)               (donne les coordonnées)
```

Chacun fait ce qu'il sait faire : **le LLM comprend, Mapbox localise.**

### Pourquoi un LLM *local* (Ollama)

Claude/OpenAI sont payants. Pour un projet perso, on fait tourner un **petit modèle gratuit sur la machine** (Ollama).

⚠️ Un gros modèle (23 Go) **fait chauffer l'ordinateur**. On utilise donc un **petit modèle (`llama3.2:1b`, ~1,3 Go)**, largement suffisant pour « trouver un lieu ». Trois garde-fous anti-chauffe :

1. **Petit modèle** (1B) au lieu d'un énorme
2. **On ne géocode que les articles NOUVEAUX** (pas les milliers déjà connus à chaque passage)
3. **Le modèle est déchargé de la mémoire** juste après usage (`keep_alive` court)

---

## 4. Le Stockage — Postgres

**Le problème.** Il faut garder les articles entre deux exécutions, et pouvoir les filtrer/trier vite.

**Ce qu'on fait.** Une table `articles` (titre, url, image, source, sujet, date, lieu + coordonnées, empreinte). Un index dédié à la requête de la carte (articles géolocalisés, par sujet, du plus récent au plus ancien).

---

## 5. Le Scheduler — tout seul, toutes les 30 min

**Le problème.** On ne veut pas lancer la récupération à la main.

**Ce qu'on fait.** Un minuteur déclenche le pipeline (récupère → géocode les nouveaux → enregistre) à intervalle régulier, et au démarrage. Il s'arrête proprement quand on coupe l'appli.

---

## 6. L'API HTTP

Le backend expose des routes JSON, consommées par la carte :

| Route | À quoi ça sert |
|---|---|
| `GET /api/health` | « est-ce que le serveur répond ? » |
| `GET /api/config` | donne au front le token Mapbox (pas codé en dur dans le JS) |
| `GET /api/topics` | la liste des sujets (le front construit ses filtres dessus) |
| `GET /api/articles?topic=health` | les articles géolocalisés, filtrables par sujet |
| `GET /*` | sert la carte (fichiers du frontend) |

---

## 7. Le Frontend — la carte

Carte **Mapbox** (React + Vite), style sobre « brutalist ».

- **Filtres par sujet** : des boutons (Tous / Technology / Business / …) rechargent les points. Le front ne connaît pas les sujets en dur : il les demande à `/api/topics`.
- **Regroupement (clustering)** :

  **Le problème.** Beaucoup d'articles tombent sur la même ville → les points se **superposent**, on n'en voit qu'un.

  **Ce qu'on fait.** Mapbox **regroupe** les points proches en une bulle numérotée. Cliquer :
  - sur un groupe qui peut s'éclater → **zoom** dessus,
  - sur un groupe « bloqué » (même ville, mêmes coordonnées) → **panneau latéral** listant *tous* ses articles.
- **Card au survol** : passer la souris sur un point isolé affiche une carte (image si dispo, sujet, source, titre, lien vers la source).

---

## 8. Le fil rouge : les interfaces

Chaque brique clé est définie par une **interface** (`Fetcher`, `Geocoder`, `Store`). Conséquence concrète :

```
Geocoder (interface — inchangée)
   ├── StaticGeocoder  (dictionnaire de 15 villes, 1ère version)
   └── LLMGeocoder     (Ollama + Mapbox, version actuelle)
```

Passer de l'un à l'autre = **une ligne** dans `main.go`. Le pipeline, l'API, la base, le front : **rien ne bouge**. C'est ça, une architecture propre.

---

## En résumé

| Brique | Problème réglé |
|---|---|
| Fetcher | récupérer des actus multi-sujets depuis l'extérieur |
| Déduplication | ne pas afficher 10 fois le même article |
| **Géocodage LLM** | **trouver le lieu même quand il n'est pas écrit mot pour mot** |
| Stockage | garder et retrouver vite les articles |
| Scheduler | tout faire tout seul, périodiquement |
| API | exposer les données proprement |
| Frontend | visualiser sur une carte lisible (filtres, regroupement, cards) |
