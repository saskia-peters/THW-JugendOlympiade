# Sicherheit & Performance

## Sicherheit

### SQL-Injection-Schutz

Alle Datenbankabfragen verwenden **parametrisierte Statements** — Benutzereingaben werden niemals in SQL-Strings eingebettet:

```go
// ❌ Verwundbar (wird nicht verwendet)
query := fmt.Sprintf("SELECT * FROM teilnehmende WHERE id = %d", id)

// ✅ Sicher (durchgehend verwendet)
db.Query("SELECT * FROM teilnehmende WHERE id = ?", id)
```

### Eingabe-Validierung

Alle Daten aus Excel werden in `backend/io/input.go` validiert, bevor sie die Datenbank berühren:

- Name darf nicht leer sein.
- Alter muss eine gültige Ganzzahl im Bereich 1–100 sein.
- `Fahrerlaubnis` muss exakt `"ja"` oder `"nein"` lauten (Groß-/Kleinschreibung ignoriert).
- `PreGroup` darf nur alphanumerische Zeichen enthalten und maximal 20 Zeichen lang sein.
- Eine PreGroup, die `max_groesse` überschreiten würde, wird abgelehnt.

### Backend-Score-Validierung

Scores werden in Go (`AssignScore`) gegen die konfigurierten Grenzen geprüft:

```go
if score < cfg.Ergebnisse.MinPunkte || score > cfg.Ergebnisse.MaxPunkte {
    return fmt.Errorf("score %d out of range [%d, %d]", score, min, max)
}
```

Die Frontend-Validierung ist nur eine UX-Hilfe — die Backend-Prüfung ist die maßgebliche Sicherheitsschranke.

### Keine Netzwerkverbindungen

Die Anwendung ist vollständig offline. Es werden keine ausgehenden Verbindungen hergestellt.

---

## Performance

### N+1-Abfragen vermieden

Statt Teilnehmende pro Gruppe in einer Schleife zu laden, wird eine einzige JOIN-Abfrage verwendet — Aggregation erfolgt in einer In-Memory-Map:

```go
// 1 Abfrage statt N
rows := db.Query(`
    SELECT g.group_id, t.*
    FROM gruppe g
    JOIN teilnehmende t ON g.teilnehmer_id = t.teilnehmer_id
    ORDER BY g.group_id
`)
```

**Auswirkung:** 32 Abfragen → 2 Abfragen (93 % Reduktion) für eine typische Veranstaltung.

### Transaktionen für Massen-Inserts

```go
tx, _ := db.Begin()
for _, p := range participants {
    tx.Exec("INSERT INTO teilnehmende ...", p.Name, p.OV, ...)
}
tx.Commit()
```

**Auswirkung:** ~10× schneller für große Datensätze.

### Datenbank-Indizes

Indizes werden in `InitDatabase()` angelegt:

```sql
CREATE INDEX IF NOT EXISTS idx_gruppe_group_id      ON gruppe(group_id);
CREATE INDEX IF NOT EXISTS idx_gruppe_teilnehmer_id ON gruppe(teilnehmer_id);
CREATE INDEX IF NOT EXISTS idx_scores_group_id      ON group_station_scores(group_id);
CREATE INDEX IF NOT EXISTS idx_scores_station_id    ON group_station_scores(station_id);
```

### Algorithmus-Komplexität

| Operation | Komplexität | Hinweis |
|-----------|-------------|---------|
| Teilnehmende-Verteilung | O(n·g) | n = Teilnehmende, g = Gruppen |
| Vorsortierung | O(n log n) | Einmalig vor der Verteilung |
| Betreuenden-Verteilung | O(b·g) | b = Betreuende |
| Auswertungsabfragen | O(n) | Einziger Aggregations-Durchlauf |

Für eine typische Veranstaltung (≤ 200 Teilnehmende, ≤ 25 Gruppen) dauern alle Operationen unter 1 Sekunde.

### Ressourcenverwaltung

- Datenbankverbindung wird in `shutdown()` via `defer db.Close()` geschlossen.
- Dateihandles werden unmittelbar nach der Verwendung geschlossen.
- PDF-Streams werden geleert und geschlossen, bevor zurückgegeben wird.
