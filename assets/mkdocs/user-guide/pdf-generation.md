# PDF-Ausgabe & Urkunden

Alle erzeugten PDFs werden in `pdf_ordner` gespeichert (Standard: `pdfdocs/`).

## Verfügbare PDFs

| PDF | Schaltfläche | Beschreibung |
|-----|-------------|--------------|
| Gruppenübersicht | **"Gruppen-PDF erstellen"** | Eine Seite je Gruppe mit Teilnehmenden + Betreuenden |
| Gruppenwertung | **"Gruppenwertung-PDF"** | Gruppen-Rankings mit Gesamtpunktzahl |
| OV-Wertung | **"Ortsverwertung-PDF"** | Ortsverband-Rankings |
| Teilnehmer-Urkunden | **"Urkunden Teilnehmende"** | Eine Urkunde je Teilnehmenden |
| OV-Urkunden | **"Urkunden OV"** | Eine Urkunde je Ortsverband |

---

## Teilnehmer-Urkunden

Jede teilnehmende Person erhält eine individuelle Urkunde mit:

- Veranstaltungsjahr und Ort
- Name und Ortsverband der Person
- Gruppennummer
- Platzierung („1. Platz", „2. Platz", …)
- Gruppenmitglieder (Text-Stil) oder Gruppenfoto (Bild-Stil)

### Stile

Gesteuert durch `urkunden_stil` in `config.toml`:

=== "text (Standard)"
    Listet alle Gruppenmitglieder in einer Tabelle unterhalb der Platzierung. Keine externen Dateien benötigt.

=== "picture"
    Bettet ein JPEG-Gruppenfoto statt der Mitgliedertabelle ein.

    Fotos in `bilder_ordner` ablegen (Standard: `pictures/`):

    ```
    pictures/
    ├── group_picture_001.jpg   ← Gruppe 1
    ├── group_picture_002.jpg   ← Gruppe 2
    └── ...
    ```

    Fehlt ein Foto, wird ein grauer Platzhalter angezeigt.

### Optionale Hintergrundvorlage

Datei ablegen unter:

```
templates/background_urkunde_teilnehmende.png
```

Fehlt die Datei, wird ein eingebautes programmatisches Layout verwendet.

**Vorlagenspezifikationen:**

| Eigenschaft | Wert |
|-------------|------|
| Format | PNG |
| Größe | A4 Hochformat — 210 × 297 mm |
| Auflösung | 2480 × 3508 px bei 300 DPI |

**Inhaltszonen (frei lassen für dynamischen Inhalt):**

| Inhalt | Ca. Y-Position |
|--------|---------------|
| Veranstaltungsjahr | ~35 mm von oben |
| Name | ~85 mm von oben |
| Ortsverband | ~105 mm von oben |
| Gruppennummer | ~125 mm von oben |
| Platzierung | ~140 mm von oben |
| Mitgliedertabelle / Foto | ab ~165 mm von oben |

Positionen in `backend/io/pdf_cert_teilnehmende.go` anpassen:

```go
pdf.SetXY(0, 85)  // zweiter Wert = Y-Position in mm
```

---

## OV-Urkunden

Jeder Ortsverband erhält eine Urkunde:

- **Sieger** — der/die OV mit der höchsten Durchschnittspunktzahl erhalten eine **Siegerurkunde** mit `templates/ov_winner_image.png`.
- **Alle anderen** — erhalten eine identische Teilnahmeurkunde (ohne Platzierung).

### Optionale Hintergrundvorlage

```
templates/background_urkunde_ovs.png
```

Gleiches Format wie die Teilnehmer-Urkunde (A4, 300 DPI PNG).

---

## Zertifikat-Layout

Feine Layoutanpassungen sind über `certificate_layout.json` / `certificate_layout.toml` im Projektstamm möglich. Diese Dateien steuern exakte Elementpositionen und sind über den eingebauten **Cert-Layout-Editor** im Admin-Bereich bearbeitbar.
