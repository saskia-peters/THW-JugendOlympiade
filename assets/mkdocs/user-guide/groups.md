# Gruppenverteilung

Nach erfolgreichem Excel-Import **"Gruppen zusammenstellen"** klicken.

## Verteilungsalgorithmus

### Teilnehmende-Verteilung

1. **Pre-Groups werden zuerst platziert** — Teilnehmende mit demselben `PreGroup`-Code kommen in eine dedizierte Gruppe.
2. **Gruppenanzahl** = `anzahlPreGroups + ceil(übrigenTeilnehmende / max_groesse)`.
3. **Restliche Teilnehmende** werden per Diversity-Score verteilt:

    | Faktor | Gewicht |
    |--------|---------|
    | Gleicher Ortsverband bereits in Gruppe | ×2,0 Malus |
    | Gleiches Geschlecht bereits in Gruppe | ×1,5 Malus |
    | Alter nah am Gruppenaltersdurchschnitt | +1,0 Malus |
    | Gruppengröße | +0,5 je Mitglied |

    Jede Person kommt in die Gruppe mit dem **niedrigsten Malus**.

### Betreuende-Verteilung (vier Phasen)

1. **Phase 1** — Personen mit Fahrerlaubnis gleichmäßig verteilen: eine Person pro Gruppe in Prioritätsreihenfolge.
2. **Phase 2** — Personen ohne Fahrerlaubnis folgen ihrem Ortsverband: bevorzugt die Gruppe, die bereits eine Person mit FL aus demselben OV hat.
3. **Phase 2b** — Neuausgleich: Personen ohne FL von der größten in die kleinste Gruppe verschieben, bis der Unterschied ≤ 1 ist.
4. **Phase 3 (Sicherheitsnetz)** — Gruppen ohne Betreuende erhalten eine Person aus der größten Gruppe.

### Warnmeldungen

Nach der Verteilung erscheinen Warnungen für:

- Gruppen **ohne jede Betreuungsperson**.
- Gruppen **ohne Person mit Fahrerlaubnis**.

Die Verteilung wird trotzdem gespeichert. Durch Hinzufügen weiterer Betreuender in der Excel-Datei und erneuten Import können die Warnungen behoben werden.

## Gruppen anzeigen

### Gruppen-Tab

**"Gruppen anzeigen"** öffnet die Tabellen-Ansicht. Jeder Tab zeigt:

- Teilnehmende mit Alter, Geschlecht, Ortsverband
- Betreuende mit Fahrerlaubnis-Status
- Gruppenstatistik (Anzahl, Geschlechterverteilung, OV-Verteilung)

### Eingabeübersicht (Ergebnismatrix)

**"Eingabeübersicht"** zeigt eine Matrix aller Gruppen × Stationen. Klick auf eine Zelle springt direkt zur Ergebniseingabe für diese Kombination.

## Gruppengröße ändern

`max_groesse` in `config.toml` anpassen, danach **"Gruppen zusammenstellen"** erneut klicken.

!!! warning
    Neuverteilung ist nur möglich, **bevor das erste Ergebnis gespeichert wurde**. Danach ist die Schaltfläche gesperrt.

## Gruppen-PDF erzeugen

**📊 Ausgabe → "Gruppen-PDF erstellen"** erzeugt ein mehrseitiges PDF (eine Seite je Gruppe) mit allen Teilnehmenden und Betreuenden. Datei wird in `pdf_ordner` gespeichert.
