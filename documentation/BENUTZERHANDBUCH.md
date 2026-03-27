# Benutzerhandbuch – Jugendolympiade Verwaltung

---

## Schnellstart

Für alle, die sofort loslegen möchten – hier der komplette Ablauf in sieben Schritten:

1. **Excel-Datei vorbereiten** – Tabellenblätter `Teilnehmende`, `Betreuende` und `Stationen` anlegen (Struktur siehe [unten](#voraussetzung-die-excel-datei-vorbereiten)).
2. **Anwendung starten** – Doppelklick auf `THW-JugendOlympiade.exe`. Wenn eine bestehende Datenbank gefunden wird, erscheint ein Dialog: *„Weiterarbeiten"* öffnet die vorhandene Datenbank, *„Neu starten"* sichert sie und legt eine neue an.
3. **Excel einlesen** – ⚙️ Daten → **„Excel einlesen"** → XLSX-Datei auswählen.
4. **Gruppen zusammenstellen** – 📝 Daten → **„Gruppen zusammenstellen"**. Die Anwendung verteilt Teilnehmende und Betreuende automatisch auf ausgewogene Gruppen. Falls nicht genügend Betreuende mit Fahrerlaubnis vorhanden sind, erscheint eine Warnmeldung.
5. **Gruppen prüfen** – 📝 Daten → **„Gruppen anzeigen"** oder **„Eingabeübersicht"** aufrufen. Optional: 📊 Ausgabe → **„Gruppen-PDF erstellen"** drucken.
6. **Ergebnisse eingeben** – 📝 Daten → **„Ergebniseingabe"** oder direkt aus der **„Eingabeübersicht"** eine Zelle anklicken. Gruppe und Station auswählen, Punktzahl eingeben, speichern.
7. **Auswerten und Urkunden erstellen** – 📊 Ausgabe → Auswertungen anzeigen und PDFs erstellen.

> **Alle erzeugten PDFs** landen im Ordner `pdfdocs/` neben der Anwendung.

---

## Voraussetzung: Die Excel-Datei vorbereiten

Bevor die Anwendung gestartet wird, muss eine Excel-Datei im Format **XLSX** bereitgestellt werden. Der Dateiname ist frei wählbar – die Datei wird beim Import über einen Dateidialog ausgewählt. Entscheidend ist ausschließlich die korrekte Struktur der Datei.

Die Datei muss **zwei Pflicht-Tabellenblätter** enthalten und kann ein **optionales drittes Blatt** besitzen:

---

### Tabellenblatt 1: `Teilnehmende`

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

### Tabellenblatt 2: `Betreuende`

Dieses Blatt enthält alle Betreuungspersonen. Die **erste Zeile ist die Kopfzeile**:

| Spalte | Pflichtfeld | Beschreibung |
|--------|-------------|--------------|
| `Name` | ✅ Ja | Name der betreuenden Person |
| `Ortsverband` | ✅ Ja | Ortsverband, dem die Person angehört |
| `Fahrerlaubnis` | ✅ Ja | `ja` oder `nein` – gibt an, ob die Person einen Führerschein besitzt |

**Hinweise:**
- Der Wert in der Spalte `Fahrerlaubnis` muss exakt `ja` oder `nein` lauten (Groß-/Kleinschreibung wird ignoriert). Andere Werte führen zu einer Fehlermeldung beim Import.
- Jede Gruppe erhält bei der automatischen Verteilung **mindestens eine Person mit Fahrerlaubnis**. Sind nicht genügend solche Personen vorhanden, erscheint nach der Gruppenverteilung eine **Warnmeldung**.
- Betreuende desselben Ortsverbands werden nach Möglichkeit in **dieselbe Gruppe** eingeteilt.

**Beispiel:**

| Name | Ortsverband | Fahrerlaubnis |
|------|-------------|---------------|
| Anna Meier | Berlin-Mitte | ja |
| Klaus Bauer | Hamburg-Nord | nein |
| Maria Koch | Hamburg-Nord | ja |

---

### Tabellenblatt 3: `Stationen` (optional)

Dieses Blatt legt die Stationsnamen für die spätere Ergebniseingabe fest. Die erste Zeile ist die Kopfzeile (`StationName` o. ä.), ab Zeile 2 steht je ein Stationsname pro Zeile.

Wird das Blatt weggelassen oder ist es leer, können Stationen später nicht benannt werden.

---

## Schritt 1: Excel-Datei importieren

1. Anwendung starten (Doppelklick auf `THW-JugendOlympiade.exe`).
2. Im Abschnitt **📝 Daten** auf **„Excel einlesen"** klicken.
3. Im Dateidialog die vorbereitete XLSX-Datei auswählen und öffnen.
4. Die Anwendung liest alle Teilnehmenden und Betreuenden ein und speichert sie in der Datenbank.

Nach erfolgreichem Import erscheint eine grüne Statusmeldung. Im Abschnitt **📝 Daten** wird die Schaltfläche **„Gruppen zusammenstellen"** aktiv.

> **Hinweis:** Ein erneuter Import ersetzt alle bestehenden Daten in der Datenbank.

---

## Schritt 2: Gruppen zusammenstellen

1. Im Abschnitt **📝 Daten** auf **„Gruppen zusammenstellen"** klicken.
2. Die Anwendung erstellt automatisch ausgewogene Gruppen.

Die automatische Gruppenverteilung berücksichtigt:
- Maximale Gruppengröße (konfigurierbar in `config.toml`, Standard: **8 Personen**)
- Personen mit demselben `PreGroup`-Code bleiben **zusammen**
- Alle übrigen Teilnehmenden werden möglichst gemischt nach Ortsverband und Geschlecht verteilt
- Betreuende werden so verteilt, dass jede Gruppe **mindestens eine Person mit Fahrerlaubnis** erhält
- Betreuende desselben Ortsverbands kommen in **dieselbe Gruppe**
- Betreuende ohne Fahrerlaubnis werden **gleichmäßig** auf die Gruppen aufgeteilt

Falls die Anzahl der Betreuenden mit Fahrerlaubnis nicht für alle Gruppen ausreicht, erscheint nach der Verteilung eine **Warnmeldung** mit der Liste der betroffenen Gruppen.

Nach der Verteilung werden die Schaltflächen **„Gruppen anzeigen"**, **„Eingabeübersicht"**, **„Ergebniseingabe"** und **„Gruppen-PDF erstellen"** aktiv. Die Auswertungs- und Urkundenschaltflächen bleiben gesperrt, bis das erste Ergebnis gespeichert wurde.

> **Wichtig:** Sobald mindestens ein Ergebnis gespeichert wurde, ist diese Schaltfläche gesperrt. So wird verhindert, dass eine neue Verteilung bestehende Ergebnisdaten unnötig unbrauchbar macht. Wenn Sie die Gruppenverteilung vor der Ergebniseingabe anpassen möchten (z. B. andere Gruppengröße in `config.toml` eintragen), klicken Sie erneut auf **„Gruppen zusammenstellen"**.

---

## Schritt 3: Gruppeneinteilung prüfen und als PDF erstellen

### Gruppen im Programm ansehen

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Gruppen anzeigen"** klicken.
3. Die Gruppen werden in Tabs dargestellt – ein Tab je Gruppe. Jede Gruppe zeigt:
   - Teilnehmende: Name, Ortsverband, Alter, Geschlecht
   - Betreuende: Name, Ortsverband, Fahrerlaubnis (✓ / –)
   - Gruppenstatistik: Teilnehmeranzahl, Durchschnittsalter, Ortsverbandsverteilung

### PDF erstellen

1. Abschnitt **📊 Ausgabe** öffnen.
2. Auf **„Gruppen-PDF erstellen"** klicken.
3. Nach kurzer Verarbeitung erscheint eine Erfolgsmeldung.

Die Datei **`Gruppeneinteilung.pdf`** wird im Ordner **`pdfdocs/`** neben der Anwendung gespeichert. Sie enthält eine Seite pro Gruppe mit der vollständigen Teilnehmenden- und Betreuendenliste sowie einer Gruppenstatistik.

---

## Schritt 4: Ergebnisse an den Stationen eingeben

Nachdem die Jugendolympiade stattgefunden hat, werden die erzielten Punktzahlen pro Gruppe und Station eingetragen.

### Ergebniseingabe

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Ergebniseingabe"** klicken.
3. Es wird eine Ansicht mit einem **Gruppen-Dropdown** angezeigt.
4. Gruppe aus dem Dropdown auswählen.
5. Für jede Station den erreichten **Punktestand** (konfigurierbar, Standard: 100–1200) eingeben.
6. Einzeln über **„Speichern"** pro Zeile oder gesammelt über **„💾 Alle Ergebnisse speichern"** speichern.
7. Nächste Gruppe auswählen und wiederholen.

> **Tipp:** Beim Wechsel zu einer anderen Gruppe mit ungespeicherten Eingaben erscheint eine Warnmeldung, die das Speichern aller Ergebnisse vor dem Wechsel anbietet.

### Eingabeübersicht

Die **„Eingabeübersicht"** zeigt auf einen Blick, welche Stationsergebnisse bereits eingetragen wurden:

1. Abschnitt **📝 Daten** → **„Eingabeübersicht"** klicken.
2. Es erscheint eine Matrix: Zeilen = Stationen, Spalten = Gruppen.
   - **Grünes ✓** – Ergebnis für diese Gruppe/Station ist gespeichert.
   - **Rotes ✗** – Ergebnis fehlt noch.
3. Auf eine beliebige Zelle klicken, um direkt zur Ergebniseingabe für diese Gruppe und Station zu springen.

---

## Schritt 5: Auswertungen ansehen

Sobald alle Ergebnisse eingetragen sind, können die Auswertungen in der Anwendung eingesehen werden.

### Auswertung nach Gruppen

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Gruppen"** klicken.
2. Es erscheint eine Rangliste aller Gruppen, sortiert nach Gesamtpunktzahl (absteigend).
3. Die Podiumsplätze (1.–3.) sind optisch hervorgehoben.
4. Darunter sind Gesamtstatistiken sichtbar: Durchschnittsergebnis, höchstes und niedrigstes Ergebnis.
5. Mit **„📄 PDF erstellen"** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Gruppe.pdf`** im Ordner `pdfdocs/` erzeugt.

### Auswertung nach Ortsverband

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Ortsverband"** klicken.
2. Es erscheint eine Rangliste aller Ortsverbände, sortiert nach Durchschnittspunktzahl.
3. Spalten: Platz, Ortsverband, Anzahl Teilnehmende, Gesamtpunkte, Ø Score.
4. Mit **„📄 PDF erstellen"** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Ortsverband.pdf`** im Ordner `pdfdocs/` erzeugt.

---

## Schritt 6: Urkunden erstellen

> **Hinweis:** Die Urkundenschaltflächen sind erst aktiv, sobald mindestens ein Ergebnis gespeichert wurde.

### Urkunden Teilnehmende

1. Abschnitt **📊 Ausgabe** → **„Urkunden Teilnehmende"** klicken.
2. Die Anwendung erzeugt für jede teilnehmende Person eine individuelle Urkunde.
3. Die Datei **`Urkunden_Teilnehmende.pdf`** wird im Ordner **`pdfdocs/`** gespeichert. Jede Seite enthält:
   - Name der Person
   - Ortsverband
   - Gruppenbezeichnung
   - Erreichter Platz der Gruppe
   - **Stil `text`** (Standard): Liste aller Gruppenmitglieder als Tabelle
   - **Stil `picture`**: Foto der Gruppe (Datei `group_picture_XXX.jpg` aus dem konfigurierten Bilderordner)

> **Hinweis:** Der Stil wird über `urkunden_stil` in `config.toml` gesteuert (`"text"` oder `"picture"`). Beim Stil `picture` werden Fotos aus dem Ordner `bilder_ordner` (Standard: `pictures/`) erwartet. Dateinamen: `group_picture_001.jpg`, `group_picture_002.jpg`, usw. Der Ordner wird beim Programmstart automatisch angelegt.

> **Hinweis:** Falls die Datei `certificate_template.png` im Programmverzeichnis liegt, wird sie als Hintergrundbild genutzt.

### Urkunden Ortsverbände

1. Abschnitt **📊 Ausgabe** → **„Urkunden Ortsverbände"** klicken.
2. Die Anwendung erzeugt für jeden Ortsverband eine eigene Seite.
3. Die Datei **`Urkunden_Ortsverbaende.pdf`** wird im Ordner **`pdfdocs/`** gespeichert:
   - Der **beste Ortsverband** (höchste Ø-Punktzahl) erhält eine **Siegerurkunde** mit Trophäenbild und der Auszeichnung „Bester Ortsverband".
   - Alle **anderen Ortsverbände** erhalten eine Teilnahmeurkunde.
   - Jede Seite enthält die Liste der Teilnehmenden des Ortsverbands.
   - Der Name der Veranstaltung wird aus der Konfiguration (`config.toml`) übernommen.

> **Hinweis:** Falls die Datei `cert_background_ov.png` im Programmverzeichnis liegt, wird sie als Hintergrundbild für OV-Urkunden genutzt. Falls die Datei `ov_winner_image.png` vorhanden ist, erscheint sie als Trophäenbild auf der Siegerurkunde.

---

## Ausgabedateien – Übersicht

Alle erzeugten PDFs werden im Unterordner **`pdfdocs/`** im Programmverzeichnis gespeichert.

| Datei | Inhalt | Erzeugt durch |
|-------|--------|---------------|
| `Gruppeneinteilung.pdf` | Alle Gruppen mit Teilnehmenden- und Betreuendenliste | „Gruppen-PDF erstellen" |
| `Auswertung_nach_Gruppe.pdf` | Gruppenrangliste nach Gesamtpunktzahl | „📄 PDF erstellen" in Gruppenauswertung |
| `Auswertung_nach_Ortsverband.pdf` | Ortsverbandsrangliste nach Ø-Score | „📄 PDF erstellen" in Ortsverbandsauswertung |
| `Urkunden_Teilnehmende.pdf` | Eine Urkunde pro Teilnehmende/r | „Urkunden Teilnehmende" |
| `Urkunden_Ortsverbaende.pdf` | Eine Urkunde pro Ortsverband (Siegerurkunde für alle Besten bei Gleichstand) | „Urkunden Ortsverbände“ |

---

## Datensicherung

- **Backup erstellen:** ⚙️ Admin → **„Datenbank sichern"** – speichert eine Sicherungskopie der Datenbank.
- **Backup wiederherstellen:** ⚙️ Admin → **„Datenbank wiederherstellen"** – stellt einen früheren Stand wieder her. Achtung: alle aktuellen Daten werden dabei überschrieben.

---

## Konfiguration (`config.toml`)

Beim ersten Start der Anwendung wird automatisch eine Datei `config.toml` im Programmverzeichnis erstellt. Diese Datei kann über **⚙️ Admin → „Konfiguration bearbeiten"** direkt in der Anwendung geändert werden oder mit einem Texteditor (z. B. Notepad) bearbeitet werden.

```toml
[veranstaltung]
name = "THW-JugendOlympiade"  # Name erscheint auf Urkunden und PDFs
jahr = 2026

[gruppen]
max_groesse = 8  # Maximale Teilnehmende pro Gruppe

[ergebnisse]
min_punkte = 100   # Kleinstes erlaubtes Ergebnis pro Station
max_punkte = 1200  # Größtes erlaubtes Ergebnis pro Station

[ausgabe]
pdf_ordner = "pdfdocs"  # Unterordner für erzeugte PDFs
db_name = "data.db"     # Dateiname der SQLite-Datenbank
urkunden_stil = "text"  # Urkunden-Stil: "text" (Mitgliedertabelle) oder "picture" (Gruppenfoto)
bilder_ordner = "pictures"  # Unterordner mit Gruppenfotos (nur bei urkunden_stil = "picture")
```

> **Hinweis:** Änderungen an Gruppengröße und Punktegrenzen werden erst nach einem Neustart der Anwendung vollständig wirksam. Änderungen am PDF-Ausgabeordner werden sofort übernommen.


Bevor die Anwendung gestartet wird, muss eine Excel-Datei im Format **XLSX** bereitgestellt werden. Der Dateiname ist frei wählbar – die Datei wird beim Import über einen Dateidialog ausgewählt. Entscheidend ist ausschließlich die korrekte Struktur der Datei.

Die Datei muss **zwei Tabellenblätter** enthalten:

---

### Tabellenblatt 1: `Teilnehmende`

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
2. Im Abschnitt **📝 Daten** auf **„Excel einlesen"** klicken.
3. Im Dateidialog die vorbereitete XLSX-Datei auswählen und öffnen.
4. Die Anwendung liest alle Teilnehmenden ein und speichert sie in der Datenbank.

Nach erfolgreichem Import erscheint eine grüne Statusmeldung. Im Abschnitt **📝 Daten** wird die Schaltfläche **„Gruppen zusammenstellen"** aktiv.

> **Hinweis:** Ein erneuter Import ersetzt alle bestehenden Daten in der Datenbank.

---

## Schritt 2: Gruppen zusammenstellen

1. Im Abschnitt **📝 Daten** auf **„Gruppen zusammenstellen"** klicken.
2. Die Anwendung erstellt automatisch ausgewogene Gruppen.

Die automatische Gruppenverteilung berücksichtigt:
- Maximale Gruppengröße (konfigurierbar in `config.toml`, Standard: **8 Personen**)
- Personen mit demselben `PreGroup`-Code bleiben **zusammen**
- Alle übrigen Teilnehmenden werden möglichst gemischt nach Ortsverband und Geschlecht verteilt

Nach der Verteilung werden die Schaltflächen **„Gruppen anzeigen"**, **„Ergebniseingabe"** und **„Gruppen-PDF erstellen"** aktiv. Die Auswertungs- und Urkundenschaltflächen bleiben gesperrt, bis das erste Ergebnis gespeichert wurde.

> **Wichtig:** Sobald mindestens ein Ergebnis gespeichert wurde, ist diese Schaltfläche gesperrt. So wird verhindert, dass eine neue Verteilung bestehende Ergebnisdaten unnötig unbrauchbar macht. Wenn Sie die Gruppenverteilung vor der Ergebniseingabe anpassen möchten (z. B. andere Gruppengröße in `config.toml` eintragen), klicken Sie erneut auf **„Gruppen zusammenstellen"**.

---

## Schritt 3: Gruppeneinteilung prüfen und als PDF erstellen

### Gruppen im Programm ansehen

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Gruppen anzeigen“** klicken.
3. Die Gruppen werden in Tabs dargestellt – ein Tab je Gruppe. Jede Gruppe zeigt Name, Ortsverband, Alter und Geschlecht aller Mitglieder.

### PDF erstellen

1. Abschnitt **📊 Ausgabe** öffnen.
2. Auf **„Gruppen-PDF erstellen"** klicken.
3. Nach kurzer Verarbeitung erscheint eine Erfolgsmeldung.

Die Datei **`Gruppeneinteilung.pdf`** wird im Ordner **`pdfdocs/`** neben der Anwendung gespeichert. Sie enthält eine Seite pro Gruppe mit der vollständigen Teilnehmendenliste und einer Gruppenstatistik.

---

## Schritt 4: Ergebnisse an den Stationen eingeben

Nachdem die Jugendolympiade stattgefunden hat, werden die erzielten Punktzahlen pro Gruppe und Station eingetragen.

1. Abschnitt **📝 Daten** öffnen.
2. Auf **„Ergebniseingabe"** klicken.
3. Es wird eine Ansicht mit einem **Gruppen-Dropdown** angezeigt.
4. Gruppe aus dem Dropdown auswählen.
5. Für jede Station den erreichten **Punktestand** (konfigurierbar, Standard: 100–1200) eingeben.
6. Einzeln über **„Speichern“** pro Zeile oder gesammelt über **„💾 Alle Ergebnisse speichern“** speichern.
7. Nächste Gruppe auswählen und wiederholen.

> **Tipp:** Beim Wechsel zu einer anderen Gruppe mit ungespeicherten Eingaben erscheint eine Warnmeldung, die das Speichern aller Ergebnisse vor dem Wechsel anbietet.

---

## Schritt 5: Auswertungen ansehen

Sobald alle Ergebnisse eingetragen sind, können die Auswertungen in der Anwendung eingesehen werden.

### Auswertung nach Gruppen

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Gruppen"** klicken.
2. Es erscheint eine Rangliste aller Gruppen, sortiert nach Gesamtpunktzahl (absteigend).
3. Die Podiumsplätze (1.–3.) sind optisch hervorgehoben.
4. Darunter sind Gesamtstatistiken sichtbar: Durchschnittsergebnis, höchstes und niedrigstes Ergebnis.
5. Mit **„📄 PDF erstellen“** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Gruppe.pdf`** im Ordner `pdfdocs/` erzeugt.

### Auswertung nach Ortsverband

1. Abschnitt **📊 Ausgabe** → **„Auswertung nach Ortsverband“** klicken.
2. Es erscheint eine Rangliste aller Ortsverbände, sortiert nach Durchschnittspunktzahl.
3. Spalten: Platz, Ortsverband, Anzahl Teilnehmende, Gesamtpunkte, Ø Score.
4. Mit **„📄 PDF erstellen“** innerhalb der Ansicht wird die Datei **`Auswertung_nach_Ortsverband.pdf`** im Ordner `pdfdocs/` erzeugt.

---

## Schritt 6: Urkunden erstellen

> **Hinweis:** Die Urkundenschaltflächen sind erst aktiv, sobald mindestens ein Ergebnis gespeichert wurde.

### Urkunden Teilnehmende

1. Abschnitt **📊 Ausgabe** → **„Urkunden Teilnehmende"** klicken.
2. Die Anwendung erzeugt für jede teilnehmende Person eine individuelle Urkunde.
3. Die Datei **`Urkunden_Teilnehmende.pdf`** wird im Ordner **`pdfdocs/`** gespeichert. Jede Seite enthält:
   - Name der Person
   - Ortsverband
   - Gruppenbezeichnung
   - Erreichter Platz der Gruppe
   - Liste aller Gruppenmitglieder

> **Hinweis:** Falls die Datei `certificate_template.png` im Programmverzeichnis liegt, wird sie als Hintergrundbild genutzt.

### Urkunden Ortsverbände

1. Abschnitt **📊 Ausgabe** → **„Urkunden Ortsverbände"** klicken.
2. Die Anwendung erzeugt für jeden Ortsverband eine eigene Seite.
3. Die Datei **`Urkunden_Ortsverbaende.pdf`** wird im Ordner **`pdfdocs/`** gespeichert:
   - Der **beste Ortsverband** (höchste Ø-Punktzahl) erhält eine **Siegerurkunde** mit dem Bild `ov_winner_image.png` und der Auszeichnung „Bester Ortsverband".
   - Alle **anderen Ortsverbände** erhalten eine einheitliche **Urkunde** mit dem Text „hat erfolgreich teilgenommen" – ohne Rangangabe.
   - Jede Seite enthält die Liste der Teilnehmenden des Ortsverbands.

> **Hinweis:** Die Urkunden Ortsverbände verwenden kein Hintergrundbild, sondern ein vollständig zentriertes Layout auf DIN-A4.

---

## Ausgabedateien – Übersicht

Alle erzeugten PDFs werden im Unterordner **`pdfdocs/`** im Programmverzeichnis gespeichert.

| Datei | Inhalt | Erzeugt durch |
|-------|--------|---------------|
| `Gruppeneinteilung.pdf` | Alle Gruppen mit Teilnehmendenlisten | „Gruppen-PDF erstellen" |
| `Auswertung_nach_Gruppe.pdf` | Gruppenrangliste nach Gesamtpunktzahl | „📄 PDF erstellen" in Gruppenauswertung |
| `Auswertung_nach_Ortsverband.pdf` | Ortsverbandsrangliste nach Ø-Score | „📄 PDF erstellen" in Ortsverbandsauswertung |
| `Urkunden_Teilnehmende.pdf` | Eine Urkunde pro Teilnehmende/r | „Urkunden Teilnehmende" |
| `Urkunden_Ortsverbaende.pdf` | Eine Urkunde pro Ortsverband (Siegerurkunde für den Besten) | „Urkunden Ortsverbände" |

---

## Datensicherung

- **Backup erstellen:** ⚙️ Admin → **„Datenbank sichern“** – speichert eine Sicherungskopie der Datenbank.
- **Backup wiederherstellen:** ⚙️ Admin → **„Datenbank wiederherstellen“** – stellt einen früheren Stand wieder her. Achtung: alle aktuellen Daten werden dabei überschrieben.

---

## Konfiguration (`config.toml`)

Beim ersten Start der Anwendung wird automatisch eine Datei `config.toml` im Programmverzeichnis erstellt. Diese Datei kann über **⚙️ Admin → „Konfiguration bearbeiten“** direkt in der Anwendung geändert werden oder mit einem Texteditor (z. B. Notepad) bearbeitet werden.

```toml
[veranstaltung]
name = "THW-JugendOlympiade 2026"  # Name erscheint auf Urkunden und PDFs
jahr = 2026

[gruppen]
max_groesse = 8  # Maximale Teilnehmende pro Gruppe

[ergebnisse]
min_punkte = 100   # Kleinstes erlaubtes Ergebnis pro Station
max_punkte = 1200  # Größstes erlaubtes Ergebnis pro Station

[ausgabe]
pdf_ordner = "pdfdocs"  # Unterordner für erzeugte PDFs
```

> **Hinweis:** Änderungen an Gruppengröße und Punktegrenzen werden erst nach einem Neustart der Anwendung vollständig wirksam. Änderungen am PDF-Ausgabeordner werden sofort übernommen.
