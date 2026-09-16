-- Script de création de la base de données PostgreSQL pour Go-Quick-Actu.
-- L'application crée ce schéma et met à jour la table "sources"
-- automatiquement au démarrage (voir db/postgres.go), ce script sert pour
-- une mise en place manuelle (psql, outil d'admin, conteneur d'init, etc.).

-- À exécuter en tant que superutilisateur / rôle disposant du droit CREATEDB,
-- connecté à une base par défaut (ex: postgres), PAS à "quickactu" elle-même.
CREATE DATABASE quickactu;

-- Se connecter ensuite à la base nouvellement créée avant de lancer la suite :
--   \c quickactu

-- Sources RSS agrégées, avec leur orientation politique indicative
-- (caractérisation simplifiée, à but de transparence).
CREATE TABLE IF NOT EXISTS sources (
    id                SERIAL PRIMARY KEY,
    name              TEXT NOT NULL UNIQUE,
    url               TEXT NOT NULL UNIQUE,
    political_leaning TEXT
);

-- Mots-clés extraits des articles par l'IA, dédupliqués.
CREATE TABLE IF NOT EXISTS keywords (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS articles (
    id           SERIAL PRIMARY KEY,
    title        TEXT NOT NULL,
    link         TEXT NOT NULL UNIQUE,
    description  TEXT,
    source_id    INTEGER REFERENCES sources(id),
    published_at TIMESTAMPTZ
);

-- Association many-to-many entre articles et mots-clés.
CREATE TABLE IF NOT EXISTS article_keywords (
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    keyword_id INTEGER NOT NULL REFERENCES keywords(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, keyword_id)
);

-- Tri par date de publication (utilisé par GET /articles).
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles (published_at DESC);

-- Filtrage par source.
CREATE INDEX IF NOT EXISTS idx_articles_source_id ON articles (source_id);

-- Données de référence : sources RSS agrégées par l'application.
-- (Doit rester aligné avec sources/sources.go, seule source de vérité ;
-- l'application les réinjecte de toute façon à chaque démarrage.)
INSERT INTO sources (name, url, political_leaning) VALUES
    ('France Info',  'https://www.franceinfo.fr/france.rss',                          'Service public - ligne éditoriale neutre'),
    ('Le Figaro',    'https://www.lefigaro.fr/rss/figaro_actualites.xml',              'Droite'),
    ('Le Monde',     'https://www.lemonde.fr/rss/une.xml',                            'Centre gauche'),
    ('Libération',   'https://www.liberation.fr/arc/outboundfeeds/rss/?outputType=xml', 'Gauche'),
    ('20 Minutes',   'https://www.20minutes.fr/feeds/rss-une.xml',                     'Généraliste - ligne éditoriale peu marquée'),
    ('L''Express',   'https://www.lexpress.fr/arc/outboundfeeds/rss/alaune.xml',       'Centre droit'),
    ('BFM TV',       'https://www.bfmtv.com/rss/news-24-7/',                           'Généraliste - ligne éditoriale souvent qualifiée de centre droit'),
    ('France 24',    'https://www.france24.com/fr/rss',                                'Service public - ligne éditoriale neutre'),
    ('RFI',          'https://www.rfi.fr/fr/rss',                                      'Service public - ligne éditoriale neutre'),
    ('Ouest-France', 'https://www.ouest-france.fr/rss-en-continu.xml',                 'Centre - presse régionale, ligne modérée')
ON CONFLICT (name) DO UPDATE SET
    url               = EXCLUDED.url,
    political_leaning = EXCLUDED.political_leaning;
