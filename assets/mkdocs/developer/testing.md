# Tests

## Test-Organisation

Tests liegen im Verzeichnis `test/` und nutzen das Standard-`testing`-Paket von Go.

```
test/
├── database_test.go      Datenbankoperationen
├── distribution_test.go  Gruppenverteilung (8 Tests)
├── evaluations_test.go   Auswertungsabfragen
├── input_test.go         Excel-Import-Validierung (10 Tests)
├── inserts_test.go       Insert-Operationen
├── models_test.go        Datenmodelle
├── queries_test.go       Abfragetests
├── scores_test.go        Score-Zuweisung
├── services_test.go      Service-Schicht
└── README.md             Ausführliche Test-Dokumentation
```

Paketbezogene Unit-Tests liegen zudem in `backend/database/` und `backend/config/`.

## Tests ausführen

```bash
# Alle Tests mit ausführlicher Ausgabe
cd test
go test -v

# Einzelnen Test ausführen
go test -v -run TestDistribution_BasicFunctionality

# Mit Coverage-Bericht
go test -cover

# Alle Pakete
go test ./...
```

Mit `just`:

```bash
just test
```

## Test-Suiten

### Excel-Import (`input_test.go`) — ~85 % Coverage

| Test | Szenario |
|------|----------|
| `TestReadXLSXFile_ValidFile` | Happy Path |
| `TestReadXLSXFile_InvalidPath` | Datei nicht gefunden |
| `TestReadXLSXFile_InvalidHeaders` | Falsche Spaltennamen |
| `TestReadXLSXFile_MissingRequiredField` | Leere Pflichtfelder |
| `TestReadXLSXFile_InvalidAge` | Alter außerhalb 1–100 |
| `TestReadXLSXFile_NonNumericAge` | Alter keine Zahl |
| `TestReadXLSXFile_EmptySheet` | Blatt ohne Datenzeilen |
| `TestValidateHeaders_Valid` | Korrekte Header (positiv) |
| `TestValidateHeaders_Invalid` | Falsche Header (negativ) |
| `TestValidateParticipantRow` | Vollständige Zeilen-Validierung |

### Gruppenverteilung (`distribution_test.go`) — ~90 % Coverage

| Test | Szenario |
|------|----------|
| `TestDistribution_BasicFunctionality` | Standardverteilung |
| `TestDistribution_EmptyParticipants` | Leere Eingabe |
| `TestDistribution_GroupSizeLimit` | Einhaltung von `max_groesse` |
| `TestDistribution_SingleParticipant` | Edge Case: 1 Person |
| `TestDistribution_ExactlyMaxSize` | Edge Case: genau 8 Personen |
| `TestDistribution_StatisticsTracking` | Statistikgenauigkeit |
| `TestDistribution_DiversityScoring` | Diversity-Algorithmus |
| `TestDistribution_ConsistentOutput` | Deterministische Ergebnisse |

## Neue Tests schreiben

```go
package test

import (
    "testing"
    "THW-JugendOlympiade/backend/services"
)

func TestNeuesFunktion_Szenario(t *testing.T) {
    // Arrange
    eingabe := testdatenVorbereiten()

    // Act
    ergebnis, err := services.NeuesFunktion(eingabe)

    // Assert
    if err != nil {
        t.Fatalf("unerwarteter Fehler: %v", err)
    }
    if ergebnis != erwartet {
        t.Errorf("erwartet %v, erhalten %v", erwartet, ergebnis)
    }
}
```

**Best Practices:**

- **Table-driven Tests** für mehrere Szenarien verwenden.
- **In-Memory-SQLite** (`:memory:` als Pfad) für DB-Tests — kein Datei-I/O.
- Aufräumen in `t.Cleanup(func() { ... })`.
- Sowohl **Happy Path** als auch **Fehlerfälle** testen.
- Benennung: `Test<Subjekt>_<Szenario>`.
