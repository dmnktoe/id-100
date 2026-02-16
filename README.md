# Innenstadt ID 100 - Companion App

Eine moderne Go-Webanwendung fuer kreative Dokumentationen mit Echo, PostgreSQL, MinIO (S3-kompatibel) und Meilisearch. Dieses Projekt ist ein nicht-kommerzielles künstlerisch-experimentelles Bachelorprojekt im Hochschulkontext der Universität Kassel. Teilnehmende laden im Rahmen von Aufgaben (Übungen zur urbanen Praxis und Innenstadtwahrnehmung) Fotos und Bilder hoch, um Wahrnehmungen des Stadtraums zu dokumentieren.

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/dmnktoe/id-100/go.yml?style=flat&logo=Github&labelColor=000&label=CI)
![GitHub Release](https://img.shields.io/github/v/release/dmnktoe/id-100?style=flat&logo=Github&labelColor=000&label=Version)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go&labelColor=000)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql&labelColor=000)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker&labelColor=000)](https://www.docker.com/)
[![Orte](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fid-100.online%2Fapi%2Fstats&query=$.total_cities&label=Orte&labelColor=000&color=9031aa)](https://id-100.online)
[![Beitraege](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fid-100.online%2Fapi%2Fstats&query=$.total_contributions&label=Beitr%C3%A4ge&labelColor=000&color=613cb1)](https://id-100.online)
[![Teilnehmer*innen](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fid-100.online%2Fapi%2Fstats&query=$.active_users&label=Teilnehmer*innen&labelColor=000&color=54b9d1)](https://id-100.online)

## Features

- Upload und Galerie fuer kreative Beitraege
- WebP-Konvertierung und LQIP fuer schnelle Bildausgabe
- S3-kompatibler Storage ueber MinIO
- PostgreSQL fuer Daten und Migrations
- Meilisearch fuer City Autocomplete
- Hot Reload via Air
- Komplettes lokales Setup via Docker Compose

## Voraussetzungen

- Go 1.24 oder hoeher
- Node.js 20 oder hoeher (Frontend Build)
- Docker und Docker Compose (empfohlen)

## Schnellstart mit Docker Compose

Lokale Entwicklung (ohne nginx):

```bash
git clone https://github.com/dmnktoe/id-100.git
cd id-100
docker compose -f docker-compose.dev.yml --env-file .env.dev up -d
```

Produktion (mit nginx):

```bash
docker compose up -d
```

Hinweise:

- In Produktion wird ein externes Netzwerk namens nginx-proxy erwartet.
- Der erste Start laedt deutsche Staedtedaten von GeoNames (ca. 10 MB).

## Manuelle Entwicklung

```bash
go mod download
npm install
```

Start:

```bash
make run
```

Build:

```bash
make build
./bin/id-100
```

## Konfiguration

Die Beispielwerte stehen in [.env.example](.env.example). Fuer die lokale Docker-Entwicklung wird [.env.dev](.env.dev) verwendet.

Wichtige Variablen:

| Variable | Beschreibung |
|---|---|
| `DATABASE_URL` | PostgreSQL Verbindung fuer App und Migrationen |
| `POSTGRES_USER` | DB Nutzer (Docker) |
| `POSTGRES_PASSWORD` | DB Passwort (Docker) |
| `POSTGRES_DB` | DB Name (Docker) |
| `S3_ACCESS_KEY` | S3 Access Key (MinIO) |
| `S3_SECRET_KEY` | S3 Secret Key (MinIO) |
| `S3_BUCKET_NAME` | Bucket fuer Uploads |
| `S3_ENDPOINT` | Interner S3 Endpoint (Container) |
| `S3_PUBLIC_URL` | Externer S3 URL fuer Browser |
| `MINIO_ROOT_USER` | MinIO Root User |
| `MINIO_ROOT_PASSWORD` | MinIO Root Password |
| `MEILI_MASTER_KEY` | Meilisearch Master Key |
| `GEOCODING_API_URL` | Meilisearch API URL |
| `SESSION_SECRET` | Session Secret |
| `ADMIN_USERNAME` | Admin User |
| `ADMIN_PASSWORD` | Admin Passwort |

## Datenbank und Migrationen

Die Migrationen liegen in [internal/database/migrations](internal/database/migrations).

- [001_create_initial_tables.sql](internal/database/migrations/001_create_initial_tables.sql) legt die Kern-Tabellen an.
- [002_insert_initial_deriven.sql](internal/database/migrations/002_insert_initial_deriven.sql) fuellt die 100 Deriven.

Wenn du eigene Deriven verwenden willst, ersetze die Inhalte von [002_insert_initial_deriven.sql](internal/database/migrations/002_insert_initial_deriven.sql) und starte den Stack neu.

## API Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/health` | Health Check (JSON) |
| `GET` | `/api/stats` | Statistik fuer Badges (JSON) |
| `GET` | `/` | Index der Deriven |
| `GET` | `/id/:number` | Detailansicht einer Derive |
| `GET` | `/upload` | Upload Formular |
| `POST` | `/upload` | Beitrag hochladen |
| `POST` | `/upload/set-name` | Spielernamen setzen |
| `POST` | `/upload/contributions/:id/delete` | Eigenen Beitrag loeschen |
| `GET` | `/leitfaden` | Leitfaden |
| `GET` | `/impressum` | Impressum |
| `GET` | `/datenschutz` | Datenschutz |
| `GET` | `/werkzeug-anfordern` | Werkzeug anfordern |
| `POST` | `/werkzeug-anfordern` | Werkzeug anfordern (Submit) |
| `GET` | `/static/*` | Statische Dateien |

## Makefile Kurzuebersicht

Die wichtigsten Targets stehen in [Makefile](Makefile):

```bash
make run
make build
make build-all
make test
make fmt
make vet
make docker-dev-up
make docker-dev-down
make docker-dev-rebuild
make docker-up
make docker-down
```

## Frontend

```bash
npm run build
npm run build:dev
npm run watch
```

Build Output liegt in [web/static/main.js](web/static/main.js).

## Testing

Das Projekt nutzt [Vitest](https://vitest.dev/) fuer TypeScript/JavaScript Tests mit umfassender Coverage.

### Tests ausfuehren

```bash
# Alle Tests ausfuehren
npm test

# Tests im Watch-Mode
npm run test:watch

# Tests mit UI
npm run test:ui

# Tests mit Coverage Report
npm run test:coverage
```

## Projektstruktur

```
id-100/
├── cmd/id-100
├── internal
├── web
├── src
├── docker-compose.yml
├── docker-compose.dev.yml
├── Dockerfile
└── Makefile
```
