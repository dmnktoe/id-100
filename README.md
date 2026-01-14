# 🏠🆔💯 Innenstadt ID 100

Eine moderne Go-Webanwendung für kreative Beiträge mit Echo-Framework, Supabase PostgreSQL und Supabase Storage.

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4.14.0-00ADD8?style=flat)](https://echo.labstack.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat)](LICENSE)

## ✨ Features

- **Upload & Gallery**: Benutzer können kreative Beiträge hochladen
- **WebP-Konvertierung**: Automatische Bildoptimierung
- **LQIP-Support**: Low-Quality Image Placeholders für schnelles Laden
- **Supabase Storage**: Sichere Cloud-Speicherung
- **Supabase PostgreSQL**: Robuste Datenpersistenz
- **Hot-Reload**: Entwicklung mit Air
- **Responsive Design**: Modernes UI mit CSS

## 📋 Voraussetzungen

- **Go**: Version 1.24 oder höher
- **Supabase Account**: Für PostgreSQL-Datenbank und Storage
- **Docker** (optional): Für lokale Entwicklungsdatenbank

## 🚀 Installation

### 1. Repository klonen

```bash
git clone https://github.com/dmnktoe/id-100.git
cd id-100
```

### 2. Dependencies installieren

```bash
go mod download
```

### 3. Entwicklungstools installieren (optional)

```bash
# Air für Hot-Reload
go install github.com/air-verse/air@latest
```

### 4. Datenbank einrichten

**Option A: Mit Docker (empfohlen für Entwicklung)**

```bash
make docker-db
```

**Option B: Lokale PostgreSQL-Installation**

```bash
createdb id100
psql id100 < schema.sql  # Falls vorhanden
```

### 5. Umgebungsvariablen konfigurieren

Erstelle eine `.env` Datei im Projektverzeichnis:

```env
# Supabase PostgreSQL Datenbank
DATABASE_URL=postgres://postgres:[DEIN-PASSWORT]@db.[DEIN-PROJEKT-REF].supabase.co:5432/postgres

# Supabase Storage Konfiguration
SUPABASE_URL=https://[DEIN-PROJEKT-REF].supabase.co
SUPABASE_ANON_KEY=dein_anon_key
SUPABASE_SERVICE_ROLE_KEY=dein_service_role_key
S3_BUCKET_NAME=id100-images
S3_ENDPOINT=https://[DEIN-PROJEKT-REF].supabase.co/storage/v1

# Lokale Entwicklung (optional)
# DATABASE_URL=postgres://dev:pass@localhost:5432/id100?sslmode=disable
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

## 🛠️ Verfügbare Makefile-Befehle

```bash
make run         # Anwendung starten
make build       # Binary erstellen
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
│   └── id-100/           # Hauptanwendung
│       ├── main.go       # Entry Point
│       ├── routes.go     # HTTP-Routen & Handler
│       ├── database.go   # DB-Verbindung
│       ├── lqip.go       # Bildplatzhalter-Generierung
│       ├── templates.go  # Template-Rendering
│       └── utils.go      # Hilfsfunktionen
├── web/
│   ├── static/           # CSS, JS, Assets
│   └── templates/        # HTML-Templates
├── tools/                # Build-Tools
├── .air.toml            # Hot-Reload Konfiguration
├── go.mod               # Go Dependencies
└── Makefile             # Build-Automatisierung
```

## 🏗️ Technologie-Stack

| Kategorie | Technologie |
|-----------|------------|
| **Backend** | Go 1.24, Echo Framework v4 |
| **Datenbank** | Supabase PostgreSQL, pgx/v5 |
| **Storage** | Supabase Storage (S3-kompatibel) |
| **Image Processing** | go-webp, LQIP |
| **Frontend** | HTML5, CSS3, Vanilla JavaScript |
| **Dev Tools** | Air (Hot-Reload), Make |

## 🔧 Konfiguration

### Air (Hot-Reload)

Die Konfiguration befindet sich in [`.air.toml`](.air.toml). Wichtige Einstellungen:

- **Port**: 8080
- **Watch-Verzeichnisse**: cmd, web
- **Delay**: 1000ms (verhindert mehrfache Neustarts)

### Templates

Templates nutzen Go's `html/template` und befinden sich in `web/templates/`:

- `layout.html` - Basis-Layout
- `components/` - Wiederverwendbare Komponenten
- Weitere Seiten-Templates

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
| `GET` | `/` | Startseite mit neuesten Beiträgen |
| `GET` | `/deriven` | Übersicht aller Deriven |
| `GET` | `/derive/:number` | Detail-Ansicht einer Derive |
| `GET` | `/upload` | Upload-Formular |
| `POST` | `/upload` | Beitrag hochladen |
| `GET` | `/spielregeln` | Spielregeln |
| `GET` | `/about` | Über das Projekt |
| `GET` | `/static/*` | Statische Dateien |
