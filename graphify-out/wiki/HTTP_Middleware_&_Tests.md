# HTTP Middleware & Tests

> 88 nodes · cohesion 0.06

## Key Concepts

- **New()** (33 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/clock/clock.go`
- **.ServeHTTP()** (26 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router.go`
- **.WriteHeader()** (22 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router.go`
- **.Header()** (21 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router.go`
- **.Error()** (17 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/apperr/apperr.go`
- **main()** (14 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/main.go`
- **WithRequestID()** (14 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/request_id.go`
- **AccessLog()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/access_log.go`
- **CORS()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/cors.go`
- **Recover()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/recover.go`
- **RequestID()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/request_id.go`
- **run()** (9 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run.go`
- **Readyz()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- **.Write()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router.go`
- **RateLimit()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit.go`
- **TestRateLimit_AllowsUpToLimitThenBlocks()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit_test.go`
- **TestRateLimit_ResetsAfterWindowAdvance()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit_test.go`
- **WriteError()** (8 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/write_error.go`
- **TestAccessLog_WritesExpectedFields()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/access_log_test.go`
- **.writeRouteError()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router.go`
- **WriteJSON()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/json.go`
- **writeRateLimitedError()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit.go`
- **TestRouter_MatchedRouteServesNormally()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router_test.go`
- **TestRouter_UnknownPathReturns404JSON()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router_test.go`
- **TestRouter_WrongMethodReturns405JSON()** (7 connections) — `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/router_test.go`
- *... and 63 more nodes in this community*

## Relationships

- [[Project Rules, Contract & SRS]] (406 shared connections)
- [[Design Taste Frontend Rules]] (74 shared connections)
- [[Clock Package & Misc Helpers]] (23 shared connections)
- [[Animate & Apple Design Skills]] (8 shared connections)
- [[Auth Login/Register Pages]] (4 shared connections)
- [[Backend Config & CORS Tests]] (2 shared connections)

## Source Files

- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/main.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/cmd/api/run_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/apperr/apperr.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/clock/clock.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/access_log.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/access_log_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/chain.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/cors.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/cors_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/health_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/json.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/json_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/rate_limit_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/recover.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/recover_test.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/request_id.go`
- `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/internal/platform/httpx/request_id_test.go`

## Audit Trail

- EXTRACTED: 162 (31%)
- INFERRED: 355 (69%)
- AMBIGUOUS: 0 (0%)

---

*Part of the graphify knowledge wiki. See [[index]] to navigate.*