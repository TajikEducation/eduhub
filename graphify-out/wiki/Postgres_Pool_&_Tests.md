# Postgres Pool & Tests

> 10 nodes · cohesion 0.36

## Key Concepts

- **.Close()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **Open()** (6 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg.go`
- **TestOpen_ConnectsAndPings()** (5 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **.Ping()** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **pg_test.go** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **fakePinger** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **testDatabaseURL()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestOpen_SkipsWithoutTestDatabaseURL()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestOpen_UnreachableHostTimesOut()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **pg.go** (1 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg.go`

## Relationships

- [[Design Taste Frontend Rules]] (39 shared connections)
- [[Project Rules, Contract & SRS]] (1 shared connections)

## Source Files

- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`

## Audit Trail

- EXTRACTED: 21 (52%)
- INFERRED: 19 (48%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*