-- Table des actualités (potentiellement géolocalisées) affichées sur la carte.
CREATE TABLE IF NOT EXISTS articles (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT        NOT NULL,
    description   TEXT,
    url           TEXT        NOT NULL,
    source        TEXT,
    published_at  TIMESTAMPTZ,

    -- Coordonnées (NULL tant que l'article n'a pas été géocodé).
    location_name TEXT,
    lat           DOUBLE PRECISION,
    lon           DOUBLE PRECISION,

    -- Empreinte unique pour la déduplication (ex: hash de l'URL).
    hash          TEXT        NOT NULL UNIQUE,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index dédié à la requête principale de la carte :
-- les articles géolocalisés, du plus récent au plus ancien.
CREATE INDEX IF NOT EXISTS idx_articles_geo
    ON articles (published_at DESC)
    WHERE lat IS NOT NULL AND lon IS NOT NULL;
