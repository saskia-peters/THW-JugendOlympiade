# Schnellstart

## Download

Aktuelle Version von der GitHub-Releases-Seite herunterladen:

| Plattform | Datei |
|-----------|-------|
| Windows | `THW-JugendOlympiade.exe` |
| macOS | `THW-JugendOlympiade.app` |
| Linux | `THW-JugendOlympiade` |

## Systemvoraussetzungen

=== "Windows"
    - Windows 10 oder neuer
    - [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) (in Windows 11 enthalten; für Windows 10 separat herunterladen)

=== "macOS"
    - macOS 10.15 (Catalina) oder neuer
    - Beim ersten Start ggf. in **Systemeinstellungen → Sicherheit** freigeben

=== "Linux"
    - Beliebige aktuelle Distribution mit WebKitGTK
    - Ubuntu/Debian: `sudo apt-get install libwebkit2gtk-4.0-dev`

## Erster Start

1. Doppelklick auf die ausführbare Datei.
2. Falls eine bestehende Datenbank (`data.db`) vorhanden ist, erscheint ein Dialog:
    - **"Weiterarbeiten"** — öffnet die vorhandene Datenbank.
    - **"Neu starten"** — sichert die alte Datenbank und beginnt neu.
3. Das Anwendungsfenster öffnet sich. Der Ausgabeordner (`pdfdocs/`) und Beispieldaten werden automatisch angelegt.

## Vollständiger Ablauf

```mermaid
flowchart TD
    A[Excel-Datei vorbereiten] --> B[Import via 'Excel einlesen']
    B --> C[Gruppen via 'Gruppen zusammenstellen' erstellen]
    C --> D{Gruppen prüfen}
    D --> E[Gruppen-PDF drucken]
    D --> F[Ergebnisse via 'Ergebniseingabe' eintragen]
    F --> G[Auswertung anzeigen]
    G --> H[PDFs & Urkunden erzeugen]
```

### Schritt 1 — Excel-Datei vorbereiten

Siehe [Excel-Import](user-guide/excel-import.md) für die erforderliche Tabellenstruktur.

### Schritt 2 — Daten importieren

1. **📝 Daten → "Excel einlesen"** klicken.
2. XLSX-Datei im Dialog auswählen.
3. Grüne Statusmeldung bestätigt den Erfolg.

### Schritt 3 — Gruppen erstellen

**"Gruppen zusammenstellen"** klicken. Der Algorithmus verteilt Teilnehmende und Betreuende automatisch auf ausgewogene Gruppen.

!!! warning "Ergebnis-Sperre"
    Sobald das erste Ergebnis gespeichert wurde, ist **"Gruppen zusammenstellen"** gesperrt. Gruppen vor der Ergebniseingabe konfigurieren.

### Schritt 4 — Ergebnisse erfassen

**"Ergebniseingabe"** öffnet die Eingabematrix. Beliebige Zelle anklicken, um direkt zur Eingabe der jeweiligen Gruppe/Station zu springen.

### Schritt 5 — Auswerten & exportieren

Aus **📊 Ausgabe** erzeugen:

- **Gruppen-PDF** — Teilnehmerlisten je Gruppe
- **Gruppenwertung-PDF** — Gruppen-Rankings
- **Ortsverwertung-PDF** — Ortsverband-Rankings
- **Urkunden Teilnehmende** — individuelle Urkunden
- **Urkunden OV** — Ortsverband-Urkunden (Sieger erhält Siegerurkunde)
