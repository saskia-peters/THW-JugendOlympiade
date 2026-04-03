# Konfigurationsreferenz

Die Konfiguration liegt in `config.toml` im Verzeichnis der ausführbaren Datei. Die Datei wird beim ersten Start **automatisch** mit sinnvollen Standardwerten angelegt.

Bearbeitbar mit jedem Texteditor oder dem eingebauten **⚙️ Config-Editor** in der Anwendung.

## Vollständige Referenz

```toml
# config.toml — Jugendolympiade-Konfiguration

[veranstaltung]
# Name erscheint auf Urkunden und PDFs
name = "THW-JugendOlympiade"

# Veranstaltungsjahr (erscheint auf Urkunden)
# Fällt auf aktuelles Jahr zurück, wenn nicht gesetzt
jahr = 2026

# Veranstaltungsort (erscheint unten auf Urkunden)
ort = "Singen am Hohentwiel"


[gruppen]
# Maximale Teilnehmendenzahl pro Gruppe
max_groesse = 8


[ergebnisse]
# Kleinstes erlaubtes Ergebnis pro Station
min_punkte = 100

# Größtes erlaubtes Ergebnis pro Station
max_punkte = 1200


[ausgabe]
# Unterordner für erzeugte PDFs (relativ zur ausführbaren Datei)
pdf_ordner = "pdfdocs"

# SQLite-Datenbankname
# Ändern, um mehrere Veranstaltungen getrennt zu halten
db_name = "data.db"

# Urkundenstil für Teilnehmer-Urkunden:
#   "text"    — Mitgliedertabelle (Standard)
#   "picture" — Gruppenfoto
urkunden_stil = "picture"

# Unterordner mit Gruppenfotos (nur bei urkunden_stil = "picture")
# Dateinamen: group_picture_001.jpg, group_picture_002.jpg, …
bilder_ordner = "pictures"
```

## Einstellungs-Kurzreferenz

| Schlüssel | Typ | Standard | Beschreibung |
|-----------|-----|---------|--------------|
| `veranstaltung.name` | string | `"THW-JugendOlympiade"` | Veranstaltungsname auf allen PDFs |
| `veranstaltung.jahr` | integer | aktuelles Jahr | Jahr auf Urkunden |
| `veranstaltung.ort` | string | `""` | Veranstaltungsort auf Urkunden |
| `gruppen.max_groesse` | integer | `8` | Max. Teilnehmende pro Gruppe |
| `ergebnisse.min_punkte` | integer | `100` | Mindest-Punktzahl |
| `ergebnisse.max_punkte` | integer | `1200` | Höchst-Punktzahl |
| `ausgabe.pdf_ordner` | string | `"pdfdocs"` | PDF-Ausgabeverzeichnis |
| `ausgabe.db_name` | string | `"data.db"` | Datenbankname |
| `ausgabe.urkunden_stil` | string | `"picture"` | Urkundenstil (`text`/`picture`) |
| `ausgabe.bilder_ordner` | string | `"pictures"` | Gruppenfotos-Verzeichnis |

## In-App-Konfigurations-Editor

Die Anwendung enthält einen eingebauten TOML-Editor. Änderungen werden vor dem Speichern validiert — ungültiges TOML wird mit einer beschreibenden Fehlermeldung abgelehnt.

## Mehrere Veranstaltungen

Für getrennte Veranstaltungen auf demselben Rechner unterschiedliche `db_name`-Werte setzen:

```toml
[ausgabe]
db_name = "data_2026_thw.db"
```
