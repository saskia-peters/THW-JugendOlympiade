# Excel-Import

Die Anwendung liest alle Veranstaltungsdaten aus einer einzigen XLSX-Datei. Die Datei muss **zwei Pflicht-Tabellenblätter** enthalten und kann ein optionales drittes Blatt besitzen.

## Tabellenblatt 1: `Teilnehmende`

Alle angemeldeten Teilnehmenden. Die **erste Zeile ist die Kopfzeile**.

| Spalte | Pflichtfeld | Beschreibung |
|--------|-------------|--------------|
| `Name` | ✅ Ja | Vor- und Nachname |
| `Ortsverband` | ✅ Ja | Lokale Gliederung |
| `Alter` | ✅ Ja | Alter als ganze Zahl (1–100) |
| `Geschlecht` | ✅ Ja | `M`, `W` oder `D` (auch ausgeschrieben) |
| `PreGroup` | ⬜ Optional | Gruppierschlüssel — Personen mit gleichem Code kommen in dieselbe Gruppe |

**PreGroup-Regeln:**

- Nur Buchstaben und Ziffern — keine Sonderzeichen oder Leerzeichen.
- Maximale Länge: 20 Zeichen.
- Eine PreGroup, die `max_groesse` überschreiten würde, wird beim Import abgelehnt.

**Beispiel:**

| Name | Ortsverband | Alter | Geschlecht | PreGroup |
|------|-------------|-------|------------|----------|
| Max Mustermann | Berlin-Mitte | 14 | M | |
| Lena Schmidt | Hamburg-Nord | 13 | W | Team1 |
| Jonas Weber | Hamburg-Nord | 15 | M | Team1 |
| Sara Yilmaz | München-Süd | 14 | W | |

---

## Tabellenblatt 2: `Betreuende`

Alle Betreuungspersonen. Die **erste Zeile ist die Kopfzeile**.

| Spalte | Pflichtfeld | Beschreibung |
|--------|-------------|--------------|
| `Name` | ✅ Ja | Name der Betreuungsperson |
| `Ortsverband` | ✅ Ja | Lokale Gliederung |
| `Fahrerlaubnis` | ✅ Ja | `ja` oder `nein` (Groß-/Kleinschreibung irrelevant) |

!!! note "Fahrerlaubnis"
    Der Verteilungsalgorithmus garantiert **mindestens eine Person mit Fahrerlaubnis pro Gruppe**. Sind nicht genug solcher Personen vorhanden, erscheint nach der Verteilung eine Warnmeldung.

**Beispiel:**

| Name | Ortsverband | Fahrerlaubnis |
|------|-------------|---------------|
| Anna Meier | Berlin-Mitte | ja |
| Klaus Bauer | Hamburg-Nord | nein |
| Maria Koch | Hamburg-Nord | ja |

---

## Tabellenblatt 3: `Stationen` (optional)

Stationsnamen für die Ergebniseingabe. Zeile 1 = Kopfzeile, ab Zeile 2 ein Stationsname pro Zeile.

---

## Import durchführen

1. **📝 Daten → "Excel einlesen"** klicken.
2. XLSX-Datei im Dateidialog auswählen.
3. Grüne Statusmeldung = Erfolg. Rote Meldung = Validierungsfehler mit Zeilenangabe.

!!! warning "Bestehende Daten werden überschrieben"
    Ein erneuter Import ersetzt **alle** Daten (inkl. Gruppen und Ergebnisse). Vorher Datenbank sichern.

## Validierungsregeln

| Regel | Fehlermeldung |
|-------|--------------|
| Name darf nicht leer sein | `row N: name is required` |
| Alter muss eine Zahl sein | `row N: age must be a number` |
| Alter muss 1–100 sein | `row N: age must be between 1 and 100` |
| Fahrerlaubnis muss `ja`/`nein` sein | `row N: fahrerlaubnis must be 'ja' or 'nein'` |
| PreGroup enthält ungültige Zeichen | `row N: pregroup contains invalid characters` |
| PreGroup zu lang | `row N: pregroup exceeds 20 characters` |

## Beispieldatei

Eine Beispieldatei mit allen drei Blättern in der korrekten Struktur wird beim ersten Start automatisch nach `example/example_data.xlsx` extrahiert.
