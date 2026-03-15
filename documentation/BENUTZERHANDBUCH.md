# Benutzerhandbuch – Jugendolympiade Verwaltung

---

## Voraussetzung: Die Excel-Datei vorbereiten

Bevor die Anwendung gestartet wird, muss eine Excel-Datei im Format **XLSX** bereitgestellt werden. Der Dateiname ist frei wählbar – die Datei wird beim Import über einen Dateidialog ausgewählt. Entscheidend ist ausschließlich die korrekte Struktur der Datei.

Die Datei muss **zwei Tabellenblätter** enthalten:

---

### Tabellenblatt 1: `Teilnehmer`

Dieses Blatt enthält alle angemeldeten Teilnehmenden. Die **erste Zeile ist die Kopfzeile** und muss exakt folgende Spaltennamen enthalten (Groß- und Kleinschreibung wird ignoriert):

| Spalte | Pflichtfeld | Beschreibung |
|--------|-------------|--------------|
| `Name` | ✅ Ja | Vor- und Nachname der teilnehmenden Person |
| `Ortsverband` | ✅ Ja | Ortsverband, dem die Person angehört |
| `Alter` | ✅ Ja | Alter als ganze Zahl (z. B. `14`) |
| `Geschlecht` | ✅ Ja | `M`, `W` oder `D` (auch ausgeschrieben möglich) |
| `PreGroup` | ⬜ Optional | Gruppierungscode – Personen mit gleichem Code werden in dieselbe Gruppe eingeteilt |

**Hinweise zur Spalte `PreGroup`:**
- Der Code darf nur **Buchstaben und Ziffern** enthalten (keine Sonderzeichen, keine Leerzeichen).
- Maximale Länge: **20 Zeichen**.
- Personen ohne Eintrag werden automatisch auf Gruppen verteilt.

**Beispiel:**

| Name | Ortsverband | Alter | Geschlecht | PreGroup |
|------|-------------|-------|------------|----------|
| Max Mustermann | Berlin-Mitte | 14 | M | |
| Lena Schmidt | Hamburg-Nord | 13 | W | Team1 |
| Jonas Weber | Hamburg-Nord | 15 | M | Team1 |
| Sara Yilmaz | München-Süd | 14 | W | |

---

### Tabellenblatt 2: `Stationen`

Dieses Blatt ist **optional**. Falls vorhanden, legt es die Stationsnamen für die spätere Ergebniseingabe fest. Die erste Zeile ist die Kopfzeile (`StationName` o. ä.), ab Zeile 2 steht je ein Stationsname pro Zeile.

Wird das Blatt weggelassen oder ist es leer, können Stationen später nicht benannt werden.

---

## Schritt 1: Excel-Datei importieren

1. Anwendung starten (Doppelklick auf `THW-JugendOlympiade.exe`).
2. Im Abschnitt **⚙️ Admin** auf **„Lade Excel Datei"** klicken.
3. Im Dateidialog die vorbereitete XLSX-Datei auswählen und öffnen.
4. Die Anwendung liest alle Teilnehmenden ein, prüft die Dateistruktur und verteilt sie automatisch auf **ausgewogene Gruppen**.

Die automatische Gruppenverteilung berücksichtigt:
- Maximale Gruppengröße von **8 Personen**
- Personen mit demselben `PreGroup`-Code bleiben **zusammen**
- Alle übrigen Teilnehmenden werden möglichst gemischt nach Ortsverband und Geschlecht verteilt

Nach erfolgreichem Import erscheint eine grüne Statusmeldung. Die Schaltflächen im Abschnitt **📝 Daten** und **📊 Ausgabe** werden aktiv.

> **Hinweis:** Ein erneuter Import ersetzt alle bestehenden Daten in der Datenbank.

---

## Schritt 2: Gruppeneinteilung prüfen und als PDF erstellen

### Gruppen im Programm ansehen

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Gruppen"** klicken.
3. Die Gruppen werden in Tabs dargestellt – ein Tab je Gruppe. Jede Gruppe zeigt Name, Ortsverband, Alter und Geschlecht aller Mitglieder.

### PDF erstellen

1. Abschnitt **📊 Ausgabe** öffnen.
2. Auf **„Gruppen-PDF erstellen"** klicken.
3. Nach kurzer Verarbeitung erscheint eine Erfolgsmeldung.

Die Datei **`Gruppeneinteilung.pdf`** wird im Ordner **`pdfdocs/`** neben der Anwendung gespeichert. Sie enthält eine Seite pro Gruppe mit der vollständigen Teilnehmerliste und einer Gruppenstatistik.

---

## Schritt 3: Ergebnisse an den Stationen eingeben

Nachdem die Jugendolympiade stattgefunden hat, werden die erzielten Punktzahlen pro Gruppe und Station eingetragen.

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Ergebniseingabe"** klicken.
3. Es wird eine tabellarische Ansicht angezeigt. Die Tabs entsprechen den einzelnen Stationen.
4. Pro Station: für jede Gruppe den erreichten **Punktestand** eintragen und speichern.

> **Tipp:** Tabs können nacheinander abgearbeitet werden – jede Station wird einzeln gespeichert.

---

## Schritt 4: Auswertungen ansehen

Sobald alle Ergebnisse eingetragen sind, können die Auswertungen in der Anwendung eingesehen werden.

### Auswertung nach Gruppen

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Gruppen"** klicken.
2. Es erscheint eine Rangliste aller Gruppen, sortiert nach Gesamtpunktzahl (absteigend).
3. Die Podiumsplätze (1.–3.) sind optisch hervorgehoben.
4. Darunter sind Gesamtstatistiken sichtbar: Durchschnittsergebnis, höchstes und niedrigstes Ergebnis.
5. Mit **„📄 Generate PDF"** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Gruppe.pdf`** im Ordner `pdfdocs/` erzeugt.

### Auswertung nach Ortsverband

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Ortsverband"** klicken.
2. Es erscheint eine Rangliste aller Ortsverbände, sortiert nach Durchschnittspunktzahl.
3. Spalten: Platz, Ortsverband, Anzahl Teilnehmende, Gesamtpunkte, Ø Score.
4. Mit **„📄 Generate PDF"** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Ortsverband.pdf`** im Ordner `pdfdocs/` erzeugt.

---

## Schritt 5: Urkunden erstellen

1. Abschnitt **📊 Ausgabe** → **„Teilnehmer-Zertifikate"** klicken.
2. Die Anwendung erzeugt für jede teilnehmende Person eine individuelle Urkunde.
3. Die Datei **`Urkunden_Teilnehmende.pdf`** wird im Ordner **`pdfdocs/`** gespeichert. Jede Seite enthält:
   - Name der Person
   - Ortsverband
   - Gruppenbezeichnung
   - Erreichter Platz der Gruppe
   - Liste aller Gruppenmitglieder

> **Hinweis:** Falls die Datei `certificate_template.png` im Programmverzeichnis liegt, wird sie als Hintergrundbild für die Urkunden genutzt.

---

## Ausgabedateien – Übersicht

Alle erzeugten PDFs werden im Unterordner **`pdfdocs/`** im Programmverzeichnis gespeichert.

| Datei | Inhalt | Erzeugt durch |
|-------|--------|---------------|
| `Gruppeneinteilung.pdf` | Alle Gruppen mit Teilnehmerlisten | „Gruppen-PDF erstellen" |
| `Auswertung_nach_Gruppe.pdf` | Gruppenrangliste nach Gesamtpunktzahl | „📄 Generate PDF" in Gruppenauswertung |
| `Auswertung_nach_Ortsverband.pdf` | Ortsverbandsrangliste nach Ø-Score | „📄 Generate PDF" in Ortsverbandsauswertung |
| `Urkunden_Teilnehmende.pdf` | Eine Urkunde pro Teilnehmende/r | „Teilnehmer-Zertifikate" |

---

## Datensicherung

- **Backup erstellen:** ⚙️ Admin → **„Backup Database"** – speichert eine Sicherungskopie der Datenbank.
- **Backup wiederherstellen:** ⚙️ Admin → **„Restore Database"** – stellt einen früheren Stand wieder her. Achtung: alle aktuellen Daten werden dabei überschrieben.
