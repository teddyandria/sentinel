-- Table des actualités (potentiellement géolocalisées) affichées sur la carte.
CREATE TABLE IF NOT EXISTS articles (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT        NOT NULL,
    description   TEXT,
    url           TEXT        NOT NULL,
    image_url     TEXT,
    source        TEXT,
    topic         TEXT        NOT NULL DEFAULT '',
    published_at  TIMESTAMPTZ,

    -- Coordonnées (NULL tant que l'article n'a pas été géocodé).
    location_name TEXT,
    lat           DOUBLE PRECISION,
    lon           DOUBLE PRECISION,

    -- Empreinte unique pour la déduplication (ex: hash de l'URL).
    hash          TEXT        NOT NULL UNIQUE,

    -- Vecteur de sens de l'article (JSON), calculé par le modèle d'embeddings.
    -- NULL tant que l'article n'a pas été indexé pour la recherche (RAG).
    embedding     TEXT,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index dédié à la requête principale de la carte : articles géolocalisés,
-- filtrés par topic, du plus récent au plus ancien.
CREATE INDEX IF NOT EXISTS idx_articles_topic_geo
    ON articles (topic, published_at DESC)
    WHERE lat IS NOT NULL AND lon IS NOT NULL;
