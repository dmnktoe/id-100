# 🏠🆔💯 Innenstadt ID 100

Eine moderne Go-Webanwendung für kreative Beiträge mit Echo-Framework, Supabase PostgreSQL und Supabase Storage.

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4.14.0-00ADD8?style=flat)](https://echo.labstack.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat)](LICENSE)

[![CI](https://github.com/dmnktoe/id-100/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/dmnktoe/id-100/actions/workflows/go.yml)

## ✨ Features

- **Upload & Gallery**: Benutzer können kreative Beiträge hochladen
- **WebP-Konvertierung**: Automatische Bildoptimierung
- **LQIP-Support**: Low-Quality Image Placeholders für schnelles Laden
- **Supabase Storage**: Sichere Cloud-Speicherung
- **Supabase PostgreSQL**: Robuste Datenpersistenz
- **Hot-Reload**: Entwicklung mit Air
- **Responsive Design**: Modernes UI mit CSS
- **City Autocomplete**: Meilisearch-Integration für intelligente Stadtauswahl
- **Docker-Compose**: Vollständige lokale Entwicklungsumgebung mit einem Befehl

## 📋 Voraussetzungen

- **Go**: Version 1.24 oder höher
- **Node.js**: Version 20 oder höher (für Frontend-Build)
- **Docker & Docker Compose**: Für die vollständige lokale Entwicklungsumgebung (empfohlen)
- **Supabase Account**: Für PostgreSQL-Datenbank und Storage (alternative zu Docker)

## 🚀 Installation

### Option 1: Mit Docker Compose (Empfohlen)

Die einfachste Methode, um die gesamte Anwendung mit allen Abhängigkeiten lokal zu starten:

```bash
# Repository klonen
git clone https://github.com/dmnktoe/id-100.git
cd id-100

# Mit Docker Compose starten
docker-compose up -d
```

Dies startet automatisch:
- **PostgreSQL** Datenbank (Port 5432)
- **MinIO** S3-kompatibler Objektspeicher (Port 9000, Console 9001)
- **Meilisearch** Suchmaschine für Stadtsuche mit GeoNames-Daten (Port 8081)
- **ID-100** Webanwendung (Port 8080)

Die Anwendung ist verfügbar unter: `http://localhost:8080`

**Hinweis**: Der erste Start lädt automatisch deutsche Städtedaten von GeoNames.org (~10MB, dauert ca. 1 Minute).

**Deriven-Daten hinzufügen**: Um die 100 Derive-Challenges zu laden, siehe [Deriven-Daten hinzufügen](docs/ADDING_DERIVEN_DATA.md). Die Datenbank-Migrationen laufen automatisch beim Start, aber die Deriven-Daten müssen manuell über das Konvertierungsskript hinzugefügt werden.

### Option 2: Manuelle Installation

#### 1. Repository klonen

```bash
git clone https://github.com/dmnktoe/id-100.git
cd id-100
```

#### 2. Dependencies installieren

```bash
go mod download
npm install
```

#### 3. Entwicklungstools installieren (optional)

```bash
# Air für Hot-Reload
go install github.com/air-verse/air@latest
```

#### 4. Datenbank einrichten

**Option A: Mit Docker (empfohlen für Entwicklung)**

```bash
make docker-db
```

**Option B: Lokale PostgreSQL-Installation**

```bash
createdb id100
psql id100 < schema.sql  # Falls vorhanden
```

#### 5. Umgebungsvariablen konfigurieren

Erstelle eine `.env` Datei im Projektverzeichnis:

**Für Docker Compose (Standard):**

```env
# App Configuration
BASE_URL=http://localhost:8080
PORT=8080
ENVIRONMENT=development

# Admin Authentication
ADMIN_USERNAME=admin
ADMIN_PASSWORD=change_me_in_production

# Database Configuration (Docker)
DATABASE_URL=postgres://dev:pass@localhost:5432/id100?sslmode=disable

# S3 Configuration (MinIO)
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET_NAME=id100-images
S3_BUCKET=id100-images
S3_REGION=us-east-1
S3_ENDPOINT=http://localhost:9000
SUPABASE_URL=http://localhost:9000

# Session Security
SESSION_SECRET=change_this_in_production_to_random_string

# Geocoding API (Meilisearch mit GeoNames-Daten)
GEOCODING_API_URL=http://localhost:8081
```

**Für Supabase (Produktion):**

```env
# Supabase PostgreSQL Datenbank
DATABASE_URL=postgres://postgres:[DEIN-PASSWORT]@db.[DEIN-PROJEKT-REF].supabase.co:5432/postgres

# Supabase Storage Konfiguration
SUPABASE_URL=https://[DEIN-PROJEKT-REF].supabase.co
SUPABASE_ANON_KEY=dein_anon_key
SUPABASE_SERVICE_ROLE_KEY=dein_service_role_key
S3_BUCKET_NAME=id100-images
S3_ENDPOINT=https://[DEIN-PROJEKT-REF].supabase.co/storage/v1

# Geocoding API (selbst gehostete Meilisearch-Instanz mit GeoNames-Daten)
GEOCODING_API_URL=https://your-meilisearch-instance.com
```

## 🎯 Verwendung

### Entwicklungsmodus (mit Hot-Reload)

```bash
air
```

### Standard-Entwicklung

```bash
make run
# oder
go run ./cmd/id-100
```

### Produktions-Build

```bash
make build
./bin/id-100
```

Die Anwendung läuft standardmäßig auf `http://localhost:8080`

## 🎨 Frontend-Entwicklung

Das Frontend verwendet TypeScript für type-sichere, modulare Client-seitige Code.

### Frontend Build

```bash
# Dependencies installieren
npm install

# TypeScript kompilieren und bundlen
npm run build

# Entwicklungsmodus (ohne Minifizierung)
npm run build:dev

# Watch-Modus (automatischer Build bei Änderungen)
npm run watch
```

### Frontend-Struktur

```
src/
├── main.ts              # Haupteinstiegspunkt
├── brand-animation.ts   # Markenanimationen
├── drawer.ts            # Drawer/Modal-Funktionalität
├── lazy-images.ts       # Lazy-Loading für Bilder
├── form-handler.ts      # Formular-Handler
└── city-autocomplete.ts # Meilisearch City Autocomplete
```

Der TypeScript-Code wird mit **esbuild** gebündelt und minifiziert in `web/static/main.js` ausgegeben.

## 🛠️ Verfügbare Befehle

### Docker Compose

```bash
# Alle Services starten
docker-compose up -d

# Logs anzeigen
docker-compose logs -f

# Services stoppen
docker-compose down

# Services neu bauen
docker-compose up -d --build

# Alle Daten löschen (Volumes)
docker-compose down -v
```

### Makefile-Befehle

```bash
make run         # Anwendung starten
make build       # Backend-Binary erstellen
make build-all   # Backend und Frontend bauen
make test        # Tests ausführen
make fmt         # Code formatieren
make vet         # Code analysieren
make docker-db   # PostgreSQL-Container starten
make docker-stop # PostgreSQL-Container stoppen
make clean       # Build-Artefakte entfernen
```

## 📁 Projektstruktur

```
id-100/
├── cmd/
│   └── id-100/
│       └── main.go           # Entry Point
├── internal/
│   ├── config/               # Konfigurationsverwaltung
│   ├── database/             # Datenbank-Verbindung & Migrations
│   ├── handlers/             # HTTP-Handler
│   │   ├── app.go           # Hauptanwendungs-Handler
│   │   ├── admin.go         # Admin-Handler
│   │   └── routes.go        # Routen-Registrierung
│   ├── middleware/           # Middleware-Funktionen
│   │   ├── auth.go          # Authentifizierung
│   │   ├── token.go         # Token-Validierung
│   │   └── session_helpers.go # Session-Hilfsfunktionen
│   ├── models/               # Datenmodelle
│   ├── templates/            # Template-Rendering
│   ├── utils/                # Hilfsfunktionen
│   │   ├── lqip.go          # Bildplatzhalter-Generierung
│   │   ├── qr.go            # QR-Code-Generierung
│   │   ├── token.go         # Token-Generierung
│   │   └── utils.go         # Allgemeine Utilities
│   └── imgutil/              # Bildverarbeitung
├── web/
│   ├── static/               # CSS, JS, Assets
│   └── templates/
│       ├── admin/           # Admin-Templates
│       ├── app/             # Hauptanwendungs-Templates
│       ├── errors/          # Fehlerseiten
│       ├── components/      # Wiederverwendbare Komponenten
│       └── layout.html      # Basis-Layout
├── tools/                    # Build-Tools
├── .air.toml                # Hot-Reload Konfiguration
├── docker-compose.yml       # Docker Compose Konfiguration
├── Dockerfile               # Docker Build Konfiguration
├── go.mod                   # Go Dependencies
└── Makefile                 # Build-Automatisierung
```

## 🏗️ Technologie-Stack

| Kategorie | Technologie |
|-----------|------------|
| **Backend** | Go 1.24, Echo Framework v4 |
| **Datenbank** | PostgreSQL 15 (Supabase oder Docker) |
| **Storage** | MinIO / Supabase Storage (S3-kompatibel) |
| **Geocoding** | Meilisearch + GeoNames.org |
| **Image Processing** | go-webp, LQIP |
| **Frontend** | HTML5, CSS3, TypeScript, esbuild |
| **Dev Tools** | Air (Hot-Reload), Docker Compose, Make |
| **Container** | Docker, Docker Compose |

## 🔧 Konfiguration

### Air (Hot-Reload)

Die Konfiguration befindet sich in [`.air.toml`](.air.toml). Wichtige Einstellungen:

- **Port**: 8080
- **Watch-Verzeichnisse**: cmd, web
- **Delay**: 1000ms (verhindert mehrfache Neustarts)

### Templates

Templates nutzen Go's `html/template` und befinden sich in `web/templates/`:

- `layout.html` - Basis-Layout
- `admin/` - Admin-Dashboard und Verwaltung
- `app/` - Hauptanwendungs-Seiten (Upload, Deriven, etc.)
- `errors/` - Fehlerseiten (Zugriff verweigert, ungültiger Token, etc.)
- `components/` - Wiederverwendbare Komponenten (Header, Footer)

## 🧪 Testing

```bash
# Alle Tests ausführen
make test

# Spezifische Tests
go test ./cmd/id-100 -v
```

## 📝 API-Endpunkte

| Methode | Pfad | Beschreibung |
|---------|------|--------------|
| `GET` | `/` | Übersicht aller IDs (Index) |
| `GET` | `/id/:number` | Detail-Ansicht einer ID |
| `GET` | `/upload` | Upload-Formular |
| `POST` | `/upload` | Beitrag hochladen |
| `GET` | `/leitfaden` | 🚨 |
| `GET` | `/static/*` | Statische Dateien |
