# Health Check Endpoints

> 29 nodes · cohesion 0.12

## Key Concepts

- **run()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run.go`
- **.Close()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **Readyz()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- **TestRun_ServesHealthzOnEphemeralPort()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **Open()** (6 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg.go`
- **Healthz()** (5 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- **TestOpen_ConnectsAndPings()** (5 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestRun_GracefulShutdownWaitsForInFlightRequest()** (5 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **health_test.go** (5 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **.Ping()** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **TestHealth_Readyz_RespectsTimeoutEvenWhenPingerIgnoresContext()** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **run_test.go** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **health.go** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- **pg_test.go** (4 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **fakePinger** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **pingWithTimeout()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- **TestHealth_Healthz_ReturnsOKWithoutDependencies()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **TestHealth_Readyz_FailingPingerReturns503WithName()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **TestHealth_Readyz_MultipleDependenciesOnlyFailingListed()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **TestHealth_Readyz_SuccessfulPingerReturns200()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- **testDatabaseURL()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestOpen_SkipsWithoutTestDatabaseURL()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestOpen_UnreachableHostTimesOut()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`
- **TestRun_ClosesPoolAfterShutdown()** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- **run.go** (3 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run.go`
- *... and 4 more nodes in this community*

## Relationships

- No strong cross-community connections detected

## Source Files

- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/pg/pg_test.go`

## Audit Trail

- EXTRACTED: 56 (49%)
- INFERRED: 59 (51%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*