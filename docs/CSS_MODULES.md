# CSS Modules Integration

## Übersicht

Alle CSS-Klassen werden jetzt automatisch gehashed, um Namespace-Konflikte zu vermeiden und Cache-Busting zu ermöglichen.

## Funktionsweise

### 1. CSS-Klassen werden gehashed

Wenn du CSS wie folgt schreibst:

```css
/* style.css */
.container {
  max-width: 1200px;
}

.button {
  padding: 10px 20px;
}
```

Werden die Klassen automatisch in gehashte Versionen umgewandelt:

```css
.styles__container___otPnC {
  max-width: 1200px;
}

.styles__button___a1B2c {
  padding: 10px 20px;
}
```

### 2. Mapping-Datei

Die Zuordnung wird in `web/static/css-modules.json` gespeichert:

```json
{
  "container": "styles__container___otPnC",
  "button": "styles__button___a1B2c"
}
```

### 3. Template-Nutzung

In Go-Templates kannst du die `cssClass` Funktion nutzen, um die gehashten Klassennamen zu verwenden:

#### Vorher (ohne CSS Modules):

```html
<div class="container">
  <button class="button primary">Klick mich</button>
</div>
```

#### Nachher (mit CSS Modules):

```html
<div class="{{cssClass "container"}}">
  <button class="{{cssClass "button"}} {{cssClass "primary"}}">Klick mich</button>
</div>
```

#### Template-Output:

```html
<div class="styles__container___otPnC">
  <button class="styles__button___a1B2c styles__primary___xYz9q">Klick mich</button>
</div>
```

## Vorteile

1. **Namespace-Isolation**: Keine Klassen-Konflikte mehr zwischen verschiedenen Komponenten
2. **Cache-Busting**: Änderungen an CSS führen zu neuen Hash-Werten
3. **Minification-friendly**: Klassennamen werden optimiert für Produktion
4. **Swiper-Integration**: Swiper CSS-Klassen werden ebenfalls gehashed (z.B. `swiper` → `styles__swiper___ThW4m`)

## Migration bestehender Templates

Um bestehende Templates zu migrieren:

1. Ersetze `class="klassenname"` mit `class="{{cssClass "klassenname"}}"`
2. Bei mehreren Klassen: `class="{{cssClass "klasse1"}} {{cssClass "klasse2"}}"`
3. Teste das Template und überprüfe die generierten Klassennamen im Browser

## Build-Prozess

```bash
npm run build
```

Dieser Befehl:
1. Baut JavaScript mit esbuild
2. Verarbeitet CSS mit PostCSS + CSS Modules
3. Generiert `css-modules.json` mit Klassen-Mappings
4. Erstellt gehashte Dateien in `web/static/dist/`
5. Generiert `manifest.json` für Asset-Pfade

## Entwicklung

Im Development-Modus (`npm run build:dev`) funktioniert alles genauso, aber ohne Minification für besseres Debugging.

## Docker

Das Dockerfile kopiert automatisch `css-modules.json` und integriert sie in den Container.

## Fallback

Wenn eine Klasse nicht in `css-modules.json` gefunden wird, gibt `cssClass` den Originalklassennamen zurück. Dies gewährleistet Abwärtskompatibilität.
