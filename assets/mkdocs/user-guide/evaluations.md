# Auswertung

Auswertungsansichten sind ab dem ersten gespeicherten Ergebnis unter **📊 Ausgabe** verfügbar.

## Gruppenwertung

Zeigt das Ranking aller Gruppen nach Gesamtpunktzahl aller Stationen (absteigend).

| Spalte | Beschreibung |
|--------|--------------|
| Platz | Position (1., 2., …) |
| Gruppe | Gruppennummer |
| Gesamtpunktzahl | Summe aller Stationsergebnisse |

Gleichstände teilen denselben Rang. Einzel-JOIN-Query über `gruppe` und `group_station_scores` — keine N+1-Abfragen.

## Ortsverband-Wertung

Rankliste der Ortsverbände nach **Durchschnittspunktzahl pro Teilnehmenden**, berechnet über alle Gruppenscores der jeweiligen Teilnehmenden.

| Spalte | Beschreibung |
|--------|--------------|
| Platz | Position |
| Ortsverband | Name |
| Teilnehmende | Anzahl eindeutiger Teilnehmender |
| Ø Punkte | Gesamtscore ÷ Teilnehmendenzahl |

!!! note "Zählweise"
    Die Teilnehmendenzahl verwendet `COUNT(DISTINCT teilnehmer_id)`, um Duplikate durch die Stationsanzahl zu vermeiden.

## PDFs erzeugen

| PDF | Schaltfläche | Dateiname |
|-----|-------------|-----------|
| Gruppenwertung | **"Gruppenwertung-PDF"** | `Ergebnisse_Gruppenwertung.pdf` |
| Ortsverband-Wertung | **"Ortsverwertung-PDF"** | `Ergebnisse_Ortsverwertung.pdf` |

Beide Dateien werden in `pdf_ordner` gespeichert (Standard: `pdfdocs/`).
