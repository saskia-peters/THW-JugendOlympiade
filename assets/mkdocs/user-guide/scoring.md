# Ergebniseingabe

## Eingabeformular öffnen

**📝 Daten → "Ergebniseingabe"** öffnet die Scoring-Ansicht.

Alternativ aus der **Eingabeübersicht**: Beliebige Zelle anklicken, um direkt zur Eingabe der jeweiligen Gruppe/Station zu springen.

## Ergebnisse eintragen

1. **Gruppe** aus dem Dropdown auswählen.
2. Die Stationstabelle lädt alle Stationen — bereits gespeicherte Ergebnisse sind vorausgefüllt.
3. Punktzahl je Station eingeben (gültiger Bereich: `min_punkte`–`max_punkte`, Standard **100–1200**).
4. **"Speichern"** klicken (oder die Einzelspeicher-Schaltfläche pro Station).

!!! info "Dirty Tracking"
    Nicht gespeicherte Änderungen werden verfolgt. Beim Gruppenwechsel mit ungespeicherten Änderungen erscheint ein Modal zum Speichern oder Verwerfen.

## Validierung

Alle Ergebnisse werden im Backend validiert:

- Werte außerhalb `[min_punkte, max_punkte]` werden abgelehnt.
- Beide Grenzen kommen aus `config.toml` — Änderungen wirken sofort ohne Neustart.

| Einstellung | Standard | Beschreibung |
|-------------|----------|--------------|
| `min_punkte` | 100 | Mindestpunktzahl pro Station |
| `max_punkte` | 1200 | Höchstpunktzahl pro Station |

Siehe [Konfiguration](../developer/configuration.md).

## Massenspeicherung

Das Formular erlaubt das Speichern aller Stationen auf einmal. Jeder Speichervorgang nutzt `INSERT OR REPLACE` — Duplikate sind durch die `UNIQUE(group_id, station_id)`-Datenbank-Constraint ausgeschlossen.
