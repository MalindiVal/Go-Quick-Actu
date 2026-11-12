# Go-Quick-Actu

API REST en Go qui **agrège l'actualité française** depuis 10 médias, **analyse chaque article avec l'IA** et produit des **synthèses multi-sources** : les faits communs, puis la façon dont chaque média traite l'information.

## Fonctionnalités

- **Agrégation RSS** de 10 médias : France Info, Le Figaro, Le Monde, Libération, 20 Minutes, L'Express, BFM TV, France 24, RFI et Ouest-France. Chacun porte une orientation éditoriale indicative.
- **Analyse IA** de chaque article : extraction de 5 mots-clés et classement dans un des 10 sujets (Politique, Économie, International, Société, Sport, Culture, Sciences & Tech, Santé, Environnement, Faits divers).
- **Synthèse multi-sources** pour un mot-clé :
  - un résumé neutre des faits communs ;
  - pour chaque média, l'angle choisi, ce qui est mis en avant ou passé sous silence, et le vocabulaire employé.
- **IA avec repli** : Claude Opus 5.5 est le modèle principal, OpenAI (`gpt-4o-mini`) prend le relais si l'appel à Claude échoue.
- **Agrégation incrémentale** : les articles déjà analysés ne sont pas renvoyés à l'IA.

## Stack technique

| Domaine | Outil |
|---|---|
| Langage | Go 1.25 |
| Serveur HTTP | [Gin](https://github.com/gin-gonic/gin) |
| Lecture RSS | [gofeed](https://github.com/mmcdole/gofeed) |
| Base de données | PostgreSQL ([lib/pq](https://github.com/lib/pq)) |
| IA | [SDK Anthropic](https://github.com/anthropics/anthropic-sdk-go) (Claude), [SDK OpenAI](https://github.com/openai/openai-go) (repli) |
| Déploiement | Docker, GitHub Actions, GitHub Container Registry |

## Architecture

```
.
├── main.go              # Point d'entrée : charge .env, connecte la base, lance le serveur
├── routes/              # Déclaration des routes HTTP
├── controllers/         # Handlers HTTP (articles, sources, sujets, synthèse, auth)
├── articledao/          # Accès aux articles en base (lecture, écriture)
├── db/                  # Connexion PostgreSQL, création du schéma, schema.sql
├── ia/                  # Analyse et synthèse (Claude + repli OpenAI)
├── rss/                 # Lecture des flux et conversion en articles
├── sources/             # Liste des médias agrégés (source de vérité)
├── topics/              # Liste des sujets (source de vérité)
└── models/              # Modèle Article
```

Au démarrage, l'application crée le schéma PostgreSQL si besoin et remplit les tables `sources` et `topics`.

## Prérequis

- [Go 1.25+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/) avec une base `quickactu`
- Une clé API [Anthropic](https://console.anthropic.com/)
- Facultatif : une clé API [OpenAI](https://platform.openai.com/), pour le repli

## Installation

```bash
git clone https://github.com/MalindiVal/Go-Quick-Actu.git
cd Go-Quick-Actu
go mod download
```

Crée la base de données. Les tables, elles, sont créées automatiquement au démarrage :

```bash
psql -U postgres -c "CREATE DATABASE quickactu;"
```

## Configuration

Copie le modèle de configuration, puis remplis les valeurs dans `.env` :

```bash
cp .env.example .env
```

| Variable | Rôle | Défaut |
|---|---|---|
| `ANTHROPIC_API_KEY` | Clé API Claude (analyse et synthèse) | **requise** |
| `OPENAI_API_KEY` | Clé API OpenAI, utilisée seulement en repli | facultative |
| `ADMIN_API_KEY` | Clé qui protège `POST /aggregate` (en-tête `X-API-Key`) | sans elle, la route est désactivée |
| `DATABASE_URL` | URL PostgreSQL complète, prioritaire sur les variables `DB_*` | — |
| `DB_HOST` | Hôte PostgreSQL | `localhost` |
| `DB_PORT` | Port | `5432` |
| `DB_USER` | Utilisateur | `postgres` |
| `DB_PASSWORD` | Mot de passe | `root` |
| `DB_NAME` | Nom de la base | `quickactu` |
| `DB_SSLMODE` | Mode SSL | `disable` |

Le fichier `.env` est ignoré par git. Ne mets jamais de vraie clé dans `.env.example` ni dans un autre fichier commité.

## Lancement

### En local

```bash
go run .
```

Le serveur écoute sur `http://localhost:8080`.

### Avec Docker

```bash
docker build -t go-quick-actu .
docker run -p 8080:8080 --env-file .env go-quick-actu
```

Ou avec l'image publiée par la CI :

```bash
docker run -p 8080:8080 --env-file .env ghcr.io/malindival/go-quick-actu:latest
```

Dans un conteneur, `localhost` désigne le conteneur lui-même. Pour joindre un PostgreSQL installé sur ta machine, utilise `DB_HOST=host.docker.internal`.

## API

| Méthode | Route | Description |
|---|---|---|
| `POST` | `/aggregate` | Lance une agrégation en arrière-plan (protégée par clé) |
| `GET` | `/articles` | Liste les articles, du plus récent au plus ancien |
| `GET` | `/sources` | Liste les médias et leur orientation |
| `GET` | `/topics` | Liste les sujets |
| `GET` | `/synthesis?keyword=…&limit=…` | Synthèse multi-sources pour un mot-clé |

### Lancer une agrégation

```bash
curl -X POST -H "X-API-Key: <ADMIN_API_KEY>" http://localhost:8080/aggregate
```

| Code | Signification |
|---|---|
| `202` | Agrégation lancée en arrière-plan. Le bilan est écrit dans les logs du serveur. |
| `401` | Clé absente ou invalide |
| `409` | Une agrégation est déjà en cours |
| `503` | `ADMIN_API_KEY` n'est pas configurée |

### Obtenir une synthèse

```bash
curl "http://localhost:8080/synthesis?keyword=climat&limit=8"
```

- `keyword` est obligatoire. La recherche ne tient pas compte des majuscules.
- `limit` est facultatif : 8 par défaut, 20 au maximum.

Exemple de réponse :

```json
{
  "keyword": "climat",
  "articles_used": [ { "title": "…", "source": "Le Monde", "topic": "Environnement", "keywords": ["…"] } ],
  "synthesis": {
    "summary": "Synthèse neutre des faits communs…",
    "perspectives": [
      {
        "source": "Le Monde",
        "political_leaning": "Centre gauche",
        "viewpoint": "Comment ce média traite l'information…"
      }
    ]
  }
}
```

## CI/CD

Le workflow [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml) :

- **à chaque push et pull request sur `main`** : vérifie le formatage (`gofmt`), lance `go vet` et compile le projet ;
- **à chaque push sur `main`** : construit l'image Docker et la publie sur GitHub Container Registry, avec les tags `latest` et `sha-<commit>`.

## À propos des orientations politiques

Les orientations attribuées aux médias sont des **caractérisations indicatives et simplifiées**, données par souci de transparence. Elles ne constituent pas un jugement définitif sur la ligne éditoriale d'un média. La synthèse s'appuie d'ailleurs sur le contenu réel des articles, et pas seulement sur ces étiquettes.
