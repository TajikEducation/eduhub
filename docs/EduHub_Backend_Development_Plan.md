# EduHub — план разработки бэкенда (v1)

**Дата:** 2026-08-24
**Статус:** согласован (design approval gate пройден), готов к реализации
**Авторы:** go-advisor (архитектурная консультация) + go-planner (детальный план) + согласование с пользователем

Источники истины, использованные при планировании: `docs/EduHub_Technical_Specification.docx` (SRS v2.1), `docs/EduHub_Functional_Requirements.md` (FR детально, v1.1), `web/lib/data.ts` и `web/lib/app-state.tsx` (текущие контракты фронта, который бэкенд должен заменить).

---

## Goal

Построить Go-бэкенд EduHub с нуля в `/Users/abdulhalim/Desktop/My-projects/eduhub/backend/` — один SOA-бинарник (`cmd/api`), 6 функциональных сервисов внутри `internal/`, PostgreSQL+PostGIS с разделением по схемам, Redis, S3. Фронт (`web/`) сейчас живёт на mock-данных `web/lib/data.ts` и должен переехать на реальный API без потери контента.

Ключевой результат первой вехи — публичный read-path каталога, на который фронт сможет переключиться, не дожидаясь auth. Дальше — авторизация, мутации с audit trail, рейтинги, коммуникации, аналитика.

---

## Constraints

**Non-negotiables (NFR EduHub):**
- API p95 ≤ 300 мс на публичных read-эндпоинтах. Каждый запрос — с явным `context.WithTimeout`.
- Uptime ≥ 99.5% → zero-downtime миграции: UP+DOWN обязательны, `CREATE INDEX CONCURRENTLY`, `nullable → backfill → SET NOT NULL`, запрет `ADD COLUMN NOT NULL` без DEFAULT.
- Mobile-first, PWA offline-кэш → ответы кэшируемые (ETag/`Cache-Control` на публичных GET).
- Двуязычность ru/tg — архитектурно, с первого дня: JSONB `{ru,tg}`, симметрично `Bi` из `web/lib/data.ts:23`.
- **Минимизация PII детей**: `auth.children` хранит только `age_group` + `status` + `institution_id`. Имя/фото ребёнка бэкенд не принимает, не хранит, не логирует — ломает текущий фронтовый `ChildLink.name`, правка фронта нужна отдельно (см. Risks).
- Пароли — argon2id, TLS everywhere, OWASP Top 10, rate limiting на публичных эндпоинтах.
- **Audit trail модерации вводится вместе с первым мутирующим эндпоинтом** (веха 3), не ретрофитом.
- **Идемпотентность мутаций** там, где повтор опасен: регистрация, публикация отзыва, отклик на вакансию, обращение работодателя.
- TDD-мандат: NO PRODUCTION CODE WITHOUT FAILING TEST FIRST. RED (падает по правильной причине) → GREEN → REFACTOR.
- Запрет прямого импорта между `internal/<service-a>` и `internal/<service-b>` — только через интерфейс-порт, объявленный на стороне потребителя.
- Роутер — stdlib `net/http.ServeMux` (Go 1.22+ pattern-routing), без сторонних фреймворков.

**Прагматичное исключение из TDD:** bootstrap-артефакты без логики (`go.mod`, `Makefile`, `docker-compose.yml`, `.env.example`, SQL-миграции) пишутся без предварительного теста — их «тест» это `goose up && goose down` и `go build`. Всё с ветвлением/расчётом/маппингом — строго через RED→GREEN.

---

## Architecture

### Раскладка каталогов

```
backend/
  cmd/
    api/main.go            — сборка зависимостей, graceful shutdown
    devseed/main.go        — сид демо-данных из web/lib/data.ts (только APP_ENV=dev)
    worker/main.go         — фоновые задачи (вехи 4-6); до вехи 4 не существует
  migrations/              — goose, нумерация сквозная по всем схемам
  internal/
    platform/              — общая инфраструктура, не знает про домены
      config/  logger/  pg/  redis/  httpx/  apperr/  idempotency/  clock/
    moderation/             — сквозной пакет audit trail + очередь модерации
    catalog/                — веха 1 (read) + 3 (write)
      domain/  usecase/  repo/postgres/  transport/http/
    auth/                   — веха 2
    reviews/                — веха 4
    communications/         — веха 5
    analytics/               — веха 6
```

Зависимости строго внутрь: `transport/http → usecase → domain`, `repo/postgres` реализует интерфейс, объявленный в `usecase`. `domain` не импортирует ни `net/http`, ни `pgx`.

### Границы схем БД и владение

Одна БД `eduhub`, схемы `auth`, `catalog`, `reviews`, `communications`, `analytics`, `moderation`, `platform`. Пишет только сервис-владелец схемы.

**Согласованные решения по схеме (design approval gate пройден):**

1. **`moderation.audit_log` — отдельная схема `moderation`**, не внутри `reviews`. Модерируются institutions/reviews/vacancies — если таблица живёт в `reviews`, `catalog` при approve пишет в чужую схему, нарушая правило владения. Владелец — пакет `internal/moderation`, остальные сервисы пишут через порт `moderation.Recorder`.

2. **`auth.children.institution_id` — обычный физический FK на `catalog.institutions(id)`**, `ON DELETE RESTRICT` (институция с привязанными детьми не может быть физически удалена — только смена `moderation_status`, soft-архивация; институции в этом продукте не удаляются физически ни при каком сценарии MVP). Это единственное согласованное исключение из правила «кросс-схемные FK только на `user_id`» — осознанно, ради целостности данных, влияющих на верификацию отзывов (FR-15). Миграционный порядок: `catalog.institutions` создаётся в вехе 1 (веха, предшествующая auth), поэтому к моменту миграции `auth.children` в вехе 2 таблица-цель уже существует — проблемы очерёдности миграций нет.

### Сквозные конвенции (обязательны во всех вехах)

**Размещение валидации:**

| Слой | Что проверяет | Пример |
|---|---|---|
| Transport (HTTP DTO / middleware) | Синтаксис: required/format/enum/диапазон, парсинг query, нормализация | `sort` ∈ {score,price_asc,price_desc,reviews}; `lat` ∈ [-90,90] |
| Usecase / service | Семантика: инварианты операции, orchestration, authz-guard, идемпотентность | `min_price ≤ max_price`; публично видны только `moderation_status='approved'` |
| Domain / entity | Чистые инварианты модели, без transport/DB | `Rating.Score ∈ [1,5]`; `Filter.Normalize()` |
| Repo / adapters | Никакой бизнес-валидации, только техническая корректность и nil-handling | маппинг `NULL → *T`, `pgx.ErrNoRows → apperr.NotFound` |

Правило разделения: если проверку можно сделать не зная состояния системы — она в transport; если нужны другие поля/БД/политика — она в usecase. Дублирование запрещено.

**Optional-поля и `nil`:** все опциональные поля контракта — `*T`, никогда `0`/`""`/`false` как «пусто». Слайсы всегда `[]T{}` в JSON (не `null`). В `slog`/метрики — никогда `*ptr` без проверки `!= nil` (хелпер `logx.PtrOrNil`).

**Параллелизм:** вехи 0-1 — без параллелизма (карточка институции — один SQL с `LATERAL json_agg`, не errgroup на 4-5 запросов — это душит пул соединений и бьёт по p95). Вехи 4-6 — `errgroup.WithContext`+`SetLimit(N)` только для разнородного I/O (БД+Redis+S3+внешний провайдер), все тесты параллельных компонентов — под `-race`.

**Ошибки:** единая таксономия `internal/platform/apperr` (`NotFound/InvalidInput/Unauthorized/Forbidden/Conflict/RateLimited/Internal`), маппинг в HTTP — ровно в одном месте (`httpx.WriteError`).

**Логирование:** `slog` JSON, `request_id` в каждой записи, запрет на телефоны/email/адреса/поисковую строку в логах.

---

## Модерация — детальный дизайн (согласовано отдельно)

Модерируются: регистрация учреждения (pending→approved/rejected), отзывы (FR-16, SLA ≤24-48ч), споры по отзывам (FR-17/FR-35, эскалация, SLA ≤72ч), верификация владельца профиля (FR-34).

**Принятые решения (MVP, простота приоритет):**
- **Один модератор** на верификацию владельца учреждения — не два независимых подтверждения. Второе подтверждение — кандидат на будущее, если возникнут инциденты захвата профиля.
- **Отказ в регистрации учреждения — без формального пути апелляции.** Учреждение правит данные и подаёт заявку заново. Явно НЕ симметрично dispute-процессу у отзывов (тот — по FR-35, обязателен).
- **Одна общая очередь модерации** на всех модераторов — без деления по региону/фильтрам.

**Общие практики, принятые к реализации (не были предметом вопроса, входят в дизайн по умолчанию):**
- **Explicit state machine**, не булевы флаги: `institution.moderation_status: pending→approved|rejected`; `review.status: pending→approved|rejected`, отдельно `disputed→resolved(kept|removed)`. Переход — только через один код-путь (transition guard в usecase), не произвольный UPDATE.
- **Claim-паттерн** вместо голого FIFO: модератор «берёт в работу» айтем (`claimed_by`, `claimed_at` — nullable колонки на очереди), таймаут (например 15 минут) возвращает заброшенный айтем в пул. Простая колонка, не отдельный lock-сервис — команда маленькая.
- **Причина reject/dispute-resolution — обязательна и структурирована**: enum (`no_license`/`duplicate`/`fake_suspected`/`other`) + свободный текст. Нужно для будущей аналитики причин отказов.
- **Приоритет очереди по флагам**, не только по времени: FR-18 анти-спам/накрутка помечает подозрительные отзывы `flagged_reason` — они всплывают выше в очереди, не ждут своей FIFO-позиции.
- **Уведомление актора о решении обязательно** (FR-20): любое решение модератора → нотификация учреждению/родителю с причиной.
- **Self-moderation guard**: если `actor_id == owner_id` модерируемой сущности — действие запрещено на уровне usecase (актуально когда роль `institution` в будущем может получить `moderator`).
- **Транзакционность**: `moderation.Recorder.Record(ctx, tx, Entry)` принимает `pgx.Tx`, не пул — запись в audit физически невозможна вне транзакции самого изменения.

---

## Step-by-step plan

### ВЕХА 0 — Платформенный скелет (задачи 1-16, полная детализация)

Цель: `go test -race ./...` зелёный, `make run` поднимает сервис, `/healthz` и `/readyz` отвечают, миграции накатываются и откатываются, CI гоняет линт+тесты+govulncheck.

**1. `backend/go.mod`, `Makefile`, `.gitignore`, `.env.example` — bootstrap**
Что: `go mod init github.com/abdulhalim/eduhub/backend` (Go 1.23+). `Makefile`: `run/test/test-race/test-integration/lint/vet/vuln/migrate-up/migrate-down/seed`. `.env.example` с `APP_ENV/HTTP_ADDR/DATABASE_URL/LOG_LEVEL` (реальный `.env` в `.gitignore`).
Приёмка: `cd backend && go build ./...` без ошибок; `git status` не показывает `.env`.

**2. `docker-compose.yml` — Postgres+PostGIS и Redis**
Что: `postgis/postgis:16-3.4` (порт 5433), `redis:7-alpine`, healthcheck, init-скрипт создаёт `eduhub_test`.
Приёмка: `docker compose up -d && docker compose ps` — оба `healthy`; `psql "$DATABASE_URL" -c "SELECT postgis_version()"` печатает версию.

**3. `internal/platform/config` — загрузка конфигурации**
RED: (а) все переменные заданы → `Load()` без ошибки; (б) `DATABASE_URL` пуст → ошибка с именем переменной; (в) `HTTP_ADDR` не задан → дефолт `:8080`.
GREEN: типизированный `Config` (`ShutdownTimeout time.Duration`, не строка), `Load() (Config, error)`.
Приёмка: `go test ./internal/platform/config/ -v` → PASS на 3 кейсах.

**4. `internal/platform/logger` — slog**
RED: (а) `New(cfg)` с `LOG_LEVEL=info` пишет JSON с ключами `time/level/msg/service`; (б) `Debug` при `info` не пишет; (в) `PtrOrNil[T](*T) any` → `nil` для nil-указателя.
Приёмка: `go test ./internal/platform/logger/ -v` → PASS.

**5. `internal/platform/apperr` — таксономия ошибок**
RED: (а) `apperr.NotFound(...)` матчится через `errors.Is(err, apperr.ErrNotFound)`; (б) обёрнутая `fmt.Errorf("repo: %w", err)` тоже матчится; (в) `apperr.Invalid` несёт `map[string]string` полей через `errors.As`.
Приёмка: `go test ./internal/platform/apperr/ -v` → PASS.

**6. `internal/platform/httpx` — request_id middleware**
RED: (а) без заголовка генерируется UUID, в `X-Request-ID` ответа и в контексте; (б) входящий заголовок сохраняется; (в) `RequestID(context.Background())` → пустая строка, не паника.
Приёмка: `go test ./internal/platform/httpx/ -run TestRequestID -v` → PASS.

**7. `httpx` — recovery + access-log middleware**
RED: (а) panic не роняет сервер, 500 + `{"error":{"code":"internal"...}}`, стек в лог; (б) access-log содержит `method/path/status/duration_ms/request_id`, НЕ содержит query-строку целиком (запрос `?q=Саидахон` не оставляет текст в логе).
Приёмка: `go test ./internal/platform/httpx/ -run 'TestRecover|TestAccessLog' -v` → PASS.

**8. `httpx` — маппинг ошибок в HTTP**
RED: таблица `NotFound→404, Invalid→400+fields, Unauthorized→401, Forbidden→403, Conflict→409, RateLimited→429+Retry-After, Internal→500` (сообщение наружу обобщённое), произвольная ошибка → 500. Везде `request_id` в теле.
Приёмка: `go test ./internal/platform/httpx/ -run TestWriteError -v` → PASS на 8 кейсах.

**9. `httpx` — JSON in/out**
RED: (а) `Content-Type: application/json; charset=utf-8`; (б) тело >1MB → `Invalid(body_too_large)`, не 500; (в) неизвестное поле → `Invalid` (`DisallowUnknownFields`); (г) битый JSON → `Invalid`, не паника; (д) пустой слайс → `[]`, не `null`.
Приёмка: `go test ./internal/platform/httpx/ -run TestJSON -v` → PASS.

**10. `internal/platform/pg` — пул pgx**
RED (build tag `integration`): (а) `pg.Open` коннектится, `Ping` проходит; (б) недостижимый URL → ошибка в пределах `ConnectTimeout`, не висит; (в) пустой `TEST_DATABASE_URL` → `t.Skip`.
Приёмка: `go test -tags=integration ./internal/platform/pg/ -v` → PASS; без docker — SKIP, не FAIL.

**11. Миграционный инструмент + миграция 00001 (расширения и схемы)**
Что: **`pressly/goose/v3`** — выбран вместо `golang-migrate`, потому что только у goose есть `-- +goose NO TRANSACTION`, необходимая для `CREATE INDEX CONCURRENTLY` (нельзя выполнить внутри транзакции, а `golang-migrate` всегда оборачивает миграцию в транзакцию).
`00001_bootstrap.sql`: UP — `CREATE EXTENSION IF NOT EXISTS postgis, pgcrypto, pg_trgm, unaccent;` + `CREATE SCHEMA auth, catalog, reviews, communications, analytics, moderation, platform;`. DOWN — `DROP SCHEMA ... CASCADE` (расширения не дропаются — комментарий в файле).
Приёмка: `make migrate-up` → `OK 00001_bootstrap.sql`; `psql -c "\dn"` → 7 схем; `make migrate-down` откатывает чисто.

**12. `httpx` — healthz/readyz**
RED: (а) `GET /healthz` → 200 без обращения к БД; (б) `GET /readyz` с падающим pinger → 503 + имя зависимости; (в) успешный pinger → 200; (г) readyz уважает таймаут 2с (spin-pinger 5с → 503, не зависание).
Приёмка: `go test ./internal/platform/httpx/ -run TestHealth -v` → PASS.

**13. `httpx` — роутер и цепочка middleware**
RED: (а) `Chain(mw1,mw2)(h)` вызывает в объявленном порядке; (б) неизвестный путь → 404 в JSON, не html; (в) неверный метод → 405.
GREEN: обёртка над `http.ServeMux` (Go 1.22 pattern-routing).
Приёмка: `go test ./internal/platform/httpx/ -run TestRouter -v` → PASS.

**14. `cmd/api/main.go` — сборка и graceful shutdown**
RED: логика в `run(ctx, cfg, deps) error`: (а) сервер стартует на `:0`, `/healthz` → 200; (б) отмена контекста при in-flight запросе (handler спит 200мс) завершается успешно в пределах `ShutdownTimeout`; (в) после shutdown пул БД закрыт.
GREEN: `signal.NotifyContext(SIGINT,SIGTERM)`, `srv.Shutdown`, `ReadHeaderTimeout`/`WriteTimeout` выставлены (защита от Slowloris).
Приёмка: `go test -race ./cmd/api/ -v` → PASS; `make run` печатает `listening addr=:8080`, Ctrl+C → `shutdown complete` без паники.

**15. CORS и rate-limit middleware**
RED: CORS — preflight с разрешённым Origin → 204+заголовки, чужой Origin → без allow-заголовка. Rate-limit — 5 разрешённых, 6-й → 429+`Retry-After`, после `clock.Advance(time.Minute)` снова 200 (часы инжектируются через `platform/clock`, без `time.Sleep` в тестах).
GREEN: in-memory token bucket по IP с TTL-очисткой. Комментарий: при мультиинстансе лимитер станет Redis-based (веха 5) — сейчас per-instance осознанно.
Приёмка: `go test -race ./internal/platform/httpx/ -run 'TestCORS|TestRateLimit' -v` → PASS.

**16. CI и линтеры**
Что: `.golangci.yml` (`errcheck/govet/staticcheck/nilerr/nilnil/bodyclose/rowserrcheck/contextcheck/exhaustive/gosec/noctx`), GitHub Actions job `backend`: `go vet` → `golangci-lint run` → `go test -race ./...` → `go test -tags=integration ./...` (сервисный контейнер postgis) → `govulncheck ./...`.
Приёмка: `make lint` → 0 issues; `make vuln` → No vulnerabilities; push в ветку → CI job зелёный.

**Критерий готовности вехи 0:** `make test-race && make lint && make vuln` зелёные; `make migrate-up && make migrate-down` идемпотентны; `/healthz` 200, `/readyz` 503 при остановленном Postgres.

---

### ВЕХА 1 — Catalog read-path (задачи 17-31, полная детализация)

Цель: публичный API каталога, на который фронт (`web/app/(site)/search/page.tsx`, `web/components/InstitCard.tsx`, `web/app/(site)/institutions/[id]/page.tsx`) переключается с mock-данных.

Контракт эндпоинтов:
- `GET /api/v1/institutions?q&type&region&area&min_price&max_price&min_rating&transport&food&verified&sort&lat&lng&radius_km&limit&cursor`
- `GET /api/v1/institutions/{id}`
- `GET /api/v1/institutions/{id}/news?limit&cursor`
- `GET /api/v1/reference/regions`

Фильтры — 1:1 с реальным фронтом (`web/app/(site)/search/page.tsx:56-116`), плюс гео (`web/lib/geo.ts` уже умеет определять координаты клиентски).

**17. Миграция 00002 — `catalog.institutions`**
Поля: `id UUID PK DEFAULT gen_random_uuid()`, `name JSONB NOT NULL`, `types TEXT[] NOT NULL`, `region TEXT NOT NULL`, `city JSONB`, `district TEXT`, `geo GEOGRAPHY(Point,4326) NOT NULL`, `license_no TEXT`, `languages TEXT[]`, `program_level TEXT[]`, `curriculum TEXT[]`, `price INT`, транспорт/питание/скидки-поля, `verified BOOL NOT NULL DEFAULT false`, `moderation_status TEXT NOT NULL DEFAULT 'pending' CHECK (moderation_status IN ('pending','approved','rejected'))`, `plan TEXT NOT NULL DEFAULT 'free'`, `founded INT`, `students_count INT`, `rating_avg NUMERIC(3,2)`, `review_count INT NOT NULL DEFAULT 0` (денормализация — заполняется вехой 4 через порт `RatingSync`, но колонки нужны сразу — фильтр `min_rating` и `sort=score` обязаны работать одним SQL-запросом), `created_at/updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`.
Приёмка: `make migrate-up` → OK; `\d catalog.institutions` показывает все колонки; `make migrate-down` откатывает чисто.

**18. Миграция 00003 — индексы `CONCURRENTLY`**
`-- +goose NO TRANSACTION`. `GIST(geo)`; `GIN(name jsonb_path_ops)`; `GIN((name->>'ru') gin_trgm_ops)` и аналогично `'tg'` (подстрочный поиск на обоих языках); `GIN(types)`; `btree(region, district)`; частичный `btree(rating_avg DESC, id) WHERE moderation_status='approved'`; `btree(price) WHERE moderation_status='approved'`.
Решение по поиску: `pg_trgm`, не `tsvector` — словаря для таджикского в Postgres нет, `to_tsvector('simple',...)` не даёт стемминга ни для одного языка, а фронт сейчас делает подстрочное совпадение (`.includes`) — trigram воспроизводит это поведение на обоих языках одинаково.
Приёмка: `make migrate-up` без ошибок; `EXPLAIN` на гео-запросе показывает `Index Scan ... gist`.

**19. Миграция 00004 — сателлитные таблицы каталога**
`catalog.institution_staff`, `catalog.achievements` (полиморфно: `owner_type CHECK IN ('institution','staff','student')` + `owner_id UUID`, индекс `(owner_type, owner_id)`, без FK — полиморфные ссылки не поддерживают FK), `catalog.institution_gallery` (S3-ключ, не URL), `catalog.institution_alumni`, `catalog.news_articles`. Все — FK на `catalog.institutions(id) ON DELETE CASCADE` (внутри своей схемы FK разрешены свободно).
Приёмка: `make migrate-up/down` чисто; 6 таблиц в схеме `catalog`.

**20. `internal/catalog/domain` — сущности и `Filter`**
RED: (а) `Filter{MinPrice:p(500),MaxPrice:p(100)}.Validate()` → ошибка поля `min_price`; (б) `Limit=0` после `Normalize()` → 20, `Limit=500` → капается на 50; (в) `Lat` без `Lng` → ошибка; (г) `RadiusKm` без координат → ошибка; (д) пустой `Filter` валиден. Конструктор `Institution` гарантирует `Gallery`/`Types` как `[]T{}`, не `nil`.
Приёмка: `go test ./internal/catalog/domain/ -v` → PASS на 6 кейсах; `go list -deps ./internal/catalog/domain | grep -E 'pgx|net/http'` → пусто.

**21. `internal/catalog/repo/postgres` — `List` без гео**
RED (integration, транзакция+rollback фикстур): 5 институций (3 approved, 1 pending, 1 rejected). (а) пустой фильтр → только 3 approved; (б) `Region="sughd"` → 1; (в) `Types=["cat_school"]` → верное подмножество; (г) `Q="гулис"` находит по подстроке в обоих языках; (д) `MinRating=4.5` не возвращает институции с `rating_avg IS NULL` (NULL — «нет данных», не «ноль»).
GREEN: интерфейс `usecase.InstitutionRepo` + реализация, строго параметризованный билдер (`$1..$n`), без конкатенации значений. Approved-only не хардкодится в репо — приходит из usecase-фильтра.
Приёмка: `go test -tags=integration ./internal/catalog/repo/postgres/ -run TestListInstitutions -v` → PASS на 5 кейсах.

**22. `repo/postgres` — гео-фильтр и сортировка**
RED: координаты Душанбе/Худжанда/Бохтара (реальные значения из `web/lib/data.ts`, поле `geo`). (а) центр Душанбе+`radius_km=10` → только душанбинские; (б) `radius_km=500` → все; (в) при заданном гео в результате есть `DistanceM *float64`, монотонно растёт; (г) без гео `DistanceM == nil` (не `0` — контракт важен); (д) `sort=price_asc` корректен, `sort=score` ставит `NULL` в конец (`NULLS LAST`).
GREEN: `ST_DWithin(geo, ST_MakePoint($lng,$lat)::geography, $radius_m)` + `ORDER BY geo <-> point` при гео-сортировке.
Приёмка: `go test -tags=integration ./internal/catalog/repo/postgres/ -run TestListGeo -v` → PASS; `EXPLAIN ANALYZE` — Index Scan по GIST.

**23. `repo/postgres` — keyset-пагинация**
RED (25 фикстур): (а) `limit=10` → 10 записей + `next_cursor`; (б) второй курсор → следующие 10 без пересечений/пропусков; (в) последняя страница → пустой `next_cursor`; (г) битый курсор → `Invalid`, не паника/500; (д) курсор с `sort=score` отвергается при `sort=price_asc` (sort зашит в курсор).
GREEN: непрозрачный base64(JSON) курсор `{sort,last_value,last_id}`, условие `WHERE (rating_avg,id) < ($v,$id)`. Keyset вместо OFFSET — детерминированность при параллельных вставках, без деградации на росте каталога.
Приёмка: `go test -tags=integration ./internal/catalog/repo/postgres/ -run TestListPagination -v` → PASS на 5 кейсах.

**24. `repo/postgres` — карточка институции одним запросом**
RED: (а) `GetByID` собирает staff/achievements/gallery/alumni одним вызовом; (б) отсутствует → `NotFound`; (в) без сотрудников → `Staff=[]`, не `nil`; (г) NULL в `license_no`/`transport_cost` → `nil`, не `""`/`0`; (д) счётчик SQL-запросов = 1 (через pgx QueryTracer — защита от N+1 в будущих рефакторингах).
GREEN: один SQL с `LEFT JOIN LATERAL (SELECT coalesce(json_agg(...),'[]') ...) ON true` на каждую коллекцию.
Приёмка: `go test -tags=integration ./internal/catalog/repo/postgres/ -run TestGetByID -v` → PASS на 5 кейсах включая счётчик=1.

**25. `internal/catalog/usecase` — сервис List/Get**
RED (фейковый репозиторий, без БД): (а) сервис всегда подставляет `Statuses=["approved"]`, даже если DTO пытается передать другое; (б) `Normalize` применён до вызова репо; (в) ошибка репо оборачивается с сохранением `errors.Is`; (г) `Get` для `pending` → `NotFound`, не `Forbidden` (не раскрываем существование немодерированной институции); (д) отменённый контекст пробрасывается в репо.
GREEN: `catalog.Service`, здесь же объявляется порт `RatingSync` (пустой пока — точка стыковки с вехой 4).
Приёмка: `go test -race ./internal/catalog/usecase/ -v` → PASS на 5 кейсах.

**26. `transport/http` — парсинг и валидация query**
RED (~12 кейсов): `sort=hack` → 400 c полем `sort`; `min_price=abc` → 400; `lat=200` → 400; `limit=-1` → 400; `type=cat_school,cat_uni` → слайс из 2; `transport=1` → `true`, `transport=` → `nil` (не `false` — отсутствие фильтра и «выключен» это разное); неизвестный query-параметр игнорируется (форвард-совместимость с фронтом).
GREEN: `parseListQuery` — только синтаксис/нормализация, семантика (`min<max`) не дублируется (уже в `domain.Filter.Validate()`).
Приёмка: `go test ./internal/catalog/transport/http/ -run TestParseListQuery -v` → PASS на 12 кейсах.

**27. `transport/http` — handler списка + DTO ответа**
RED (httptest, фейковый usecase): (а) 200, `{"items":[...],"next_cursor":"...","total_hint":null}`; (б) 0 результатов → `"items":[]`, не `null`; (в) `distance_m` отсутствует при `nil` (`omitempty`), присутствует при заданном; (г) `NotFound` → 404 через `httpx.WriteError`; (д) `Cache-Control: public, max-age=60`.
GREEN: явный DTO-маппинг domain→response (доменные структуры не отдаются напрямую).
Приёмка: `go test ./internal/catalog/transport/http/ -run TestListHandler -v` → PASS.

**28. `transport/http` — handler карточки + ETag**
RED: (а) валидный UUID → 200 полная карточка; (б) невалидный UUID → 400, не 500; (в) ETag от `updated_at`+id, повтор с `If-None-Match` → 304 пустое тело; (г) 404 для неизвестного id.
Приёмка: `go test ./internal/catalog/transport/http/ -run TestGetHandler -v` → PASS на 4 кейсах.

**29. `cmd/api` — регистрация маршрутов + сквозной smoke-тест**
RED: реальный usecase поверх фейкового репо, `GET /api/v1/institutions` и `GET /api/v1/institutions/{id}`, проверка цепочки middleware (`X-Request-ID` в ответе), 404 на неизвестном пути в JSON.
Приёмка: `go test -race ./cmd/api/ -v` → PASS; `curl -s localhost:8080/api/v1/institutions | jq '.items | length'`.

**30. `cmd/devseed` — перенос демо-данных из `web/lib/data.ts`**
Что: сидер переносит 9 институций со всеми полями/персоналом/достижениями/галереей/alumni/новостями (`INSTITUTIONS`, `ALL_STAFF` в `web/lib/data.ts`). Отказывает при `APP_ENV != dev`. Идемпотентен (upsert по стабильному `seed_ref` в `platform.seed_refs`).
RED (integration): (а) после первого прогона — 9 approved-институций; (б) повторный прогон — по-прежнему 9; (в) `APP_ENV=prod` → ошибка, ничего не записано.
Приёмка: `go test -tags=integration ./cmd/devseed/ -v` → PASS; `make seed && curl ... | jq '.items|length'` → `9`.

**31. Redis-кэш списка + singleflight + проверка p95**
RED (in-memory фейк кэша): (а) второй идентичный запрос не доходит до репо; (б) разные фильтры → разные ключи; (в) 50 параллельных одинаковых запросов при холодном кэше → репо вызван 1 раз (singleflight, под `-race`); (г) ошибка Redis не ломает запрос — деградация к прямому чтению из БД (кэш не точка отказа).
GREEN: декоратор `CachedService` поверх `Service`, ключ `catalog:list:v{version}:{sha256(normalized_filter)}`, TTL 60с, версия из `catalog:version` (инкрементируется записью в вехе 3).
Приёмка: `go test -race ./internal/catalog/usecase/ -run TestCache -v` → PASS; `hey -z 30s -c 50 'http://localhost:8080/api/v1/institutions?region=dushanbe'` → p95 ниже 300мс, зафиксировать в `backend/docs/perf-baseline.md`.

**Критерий готовности вехи 1:** все тесты зелёные под `-race`; `hey` подтверждает p95≤300мс на сидовых данных; фронт может заменить `useVisibleInstitutions()` (`web/lib/app-state.tsx`) на fetch к API без потери отображаемых полей; линт и govulncheck чистые.

---

### ВЕХА 2 — Auth / RBAC (эпики)

**E2.1 Схема `auth`.** `users` (`phone` уникальный нормализованный `+992XXXXXXXXX`, `email` nullable уникальный, `password_hash`, `role`, `status`, `failed_login_count`, `locked_until`), `refresh_tokens` (`token_hash`, `family_id`, `expires_at`, `revoked_at`, `replaced_by`), `children` (`user_id`, `institution_id` **FK на `catalog.institutions` ON DELETE RESTRICT** — см. Architecture, `age_group`, `status ∈ {current,alumnus,transferred}`, БЕЗ имени/фото).

**E2.2 Хеширование и парольная политика.** argon2id, параметры в конфиге, PHC-формат хранения (параметры можно поднять без слома существующих хешей).

**E2.3 JWT и ротация refresh.** Access 15 мин, refresh 30 дней, хранится только хеш. **Reuse detection**: предъявление уже использованного refresh отзывает всю `family_id` + запись в audit — с самого начала, не «потом». HS256 достаточен пока валидатор один (монолит); переход на RS256/JWKS — если появится второй потребитель токенов.

**E2.4 Эндпоинты.** `POST /auth/register` (идемпотентен: повтор с тем же телефоном → 409 `phone_taken`, при `Idempotency-Key` — повтор исходного ответа), `POST /auth/login` (единое сообщение об ошибке для «нет юзера» и «неверный пароль» — anti-enumeration), `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`.

**E2.5 Middleware RBAC.** `RequireAuth` кладёт `auth.Principal{UserID,Role}` в контекст; аксессор `FromContext(ctx)(Principal,bool)` — обязательно с `ok`. `RequireRole(roles...)`. Табличный тест матрицы «роль × эндпоинт» — растёт в вехах 3-6.

**E2.6 Children CRUD с PII-минимизацией.** API принимает только `institution_id`, `age_group`, `status`. Явный тест: поле `name` в запросе → 400 `unknown field`. Проверка существования+approved-статуса институции — теперь просто обычный FK-constraint (упрощено по решению из Architecture — не нужен отдельный порт `catalog.InstitutionExists`, БД сама гарантирует).

**E2.7 Anti-bruteforce.** Rate limit на `/auth/login` по телефону и IP, экспоненциальная задержка, блокировка на 15 мин после N попыток; события в audit без пароля и с маскированным телефоном (`+992 *** ** 12`).

**Критерии готовности:** матричный тест ролей зелёный; тест reuse-detection refresh-токена зелёный; ни один тест не пишет PII в лог (grep-проверка тестового вывода); `-race` чистый.

---

### ВЕХА 3 — Catalog write + RBAC-защита мутаций + сквозной audit

**E3.1 Пакет `internal/moderation`.** Схема `moderation.audit_log` (`actor_id`, `actor_role`, `action`, `target_type`, `target_id`, `reason`, `payload_diff JSONB`, `request_id`, `created_at`). Порт `moderation.Recorder.Record(ctx, tx, Entry) error` — принимает транзакцию (запись audit физически невозможна вне транзакции изменения). Здесь же — очередь модерации с claim-паттерном (`claimed_by`, `claimed_at`, `priority`, `flagged_reason`) — см. раздел «Модерация — детальный дизайн» выше.

**E3.2 Пакет `internal/platform/idempotency`.** `platform.idempotency_keys` (`key`,`user_id`,`endpoint`,`request_hash`,`response_status`,`response_body`,`created_at`, UNIQUE(`key`,`user_id`,`endpoint`)). Тот же ключ+тело → сохранённый ответ; тот же ключ+другое тело → 422 `idempotency_key_reuse`. TTL 24ч. Переиспользуется в вехах 4-5.

**E3.3 Регистрация институции.** `POST /api/v1/institutions` (роль `user`/`institution`), `moderation_status='pending'`, владелец в `catalog.institution_owners`. Идемпотентно. Форма — минимум из `RegisterInstitutionInput` (`web/lib/app-state.tsx:31+`), остальное — дефолты.

**E3.4 Редактирование профиля.** `PATCH /api/v1/institutions/{id}` — владелец или модератор. Оптимистическая блокировка (`version INT` / `If-Match`) — конкурентная правка → 409, не «последний победил». Частичное обновление: `nil` = не трогать, явный JSON `null` = очистить (нужен custom-unmarshal, отдельная задача при реализации).

**E3.5 Модерация.** `POST /api/v1/moderation/institutions/{id}/approve|reject` (роль `moderator`/`admin`, обязательный структурированный `reason` для reject — без апелляции, см. раздел выше), `GET /api/v1/moderation/queue` (одна общая очередь, claim-эндпоинт). Каждое действие — транзакция {смена статуса + audit + инвалидация кэша + уведомление актору}.

**E3.6 Медиа.** `POST /api/v1/institutions/{id}/media/presign` → presigned PUT в S3. Бэкенд хранит только ключ. Allow-list content-type, лимит размера, имя генерируется на сервере (защита path traversal). Антивирус-сканирование — вне MVP, известный пробел.

**E3.7 Инвалидация кэша.** Запись инкрементирует `catalog:version` в Redis — после approve список отдаёт новую институцию без ожидания TTL.

**Критерии готовности:** ни один мутирующий эндпоинт не проходит тест без записи в `moderation.audit_log` (policy-тест по всем зарегистрированным мутирующим роутам); повторный POST с тем же `Idempotency-Key` не создаёт дубль.

---

### ВЕХА 4 — Reviews / Ratings

**E4.1 Схема `reviews`.** `reviews` (UNIQUE(`user_id`,`institution_id`,`child_id`) — идемпотентность на уровне БД), `review_metrics` (нормализовано по строкам: `review_id`,`metric_key`,`score 1-5`, UNIQUE(`review_id`,`metric_key`)), `institution_rating_agg` (`institution_id`,`metric_key`,`weighted_avg`,`review_count`,`updated_at`).
**8 метрик** (решено ранее в разговоре — по `docs/EduHub_Functional_Requirements.md` v1.1, уточнённая версия; девятая величина в SRS/бизнес-плане — вычисляемое среднее, не отдельная оцениваемая метрика): `quality, conditions, safety, food, transport, price, parent_involvement, inclusivity`.

**E4.2 Eligibility (верификация отзыва).** Разрешено только при наличии `auth.children` со связью user→institution в статусе `current`/`alumnus`/`transferred`. Проверка — в usecase через порт `auth.ChildLinkExists` (кросс-схемная логика — не через FK, поскольку `reviews` не владеет схемой `auth`, порт остаётся правильным паттерном в этом направлении, в отличие от `children→institutions`).

**E4.3 Recency decay.** `w_i = 0.5 ^ (age_days / H)`, `H=365` дней (конфиг, не хардкод). `weighted_avg = Σ(w_i·score_i) / Σ(w_i)` по метрике. Порог публикации: `review_count >= MIN_REVIEWS` (старт 5), иначе `rating: null` + `rating_status: "insufficient"`. Два триггера пересчёта: (1) точечный по событию «отзыв одобрен/отклонён», (2) обязательный ночной полный пересчёт (веса меняются от одного лишь течения времени).

**E4.4 Outbox + релей.** `reviews.outbox` (`id`,`topic`,`payload`,`created_at`,`published_at`). Публикация отзыва = {отзыв+метрики+outbox-запись} в одной транзакции. Релей — `FOR UPDATE SKIP LOCKED` батчами → Redis Stream → консьюмер идемпотентен по `event_id`.

**E4.5 Синхронизация с каталогом.** Консьюмер вызывает порт `catalog.RatingSync.Apply(ctx, instID, avg, count)` — `catalog` сам пишет в свои колонки, бампает cache version. Владение схемами соблюдено.

**E4.6 Фоновые задачи и лидерство.** Ночной пересчёт и чистка idempotency-ключей — в `cmd/worker`, лидер через `pg_try_advisory_lock` — безопасно на нескольких инстансах без внешнего оркестратора.

**E4.7 Модерация отзывов и споры.** Статусы `pending/approved/rejected/disputed`, ответ учреждения (`reply`), формальный спор с SLA 72ч (`dispute_deadline`, джоб-эскалация) — здесь путь апелляции ЕСТЬ (в отличие от отказа в регистрации учреждения, см. раздел «Модерация» выше). Каждое действие — в `moderation.audit_log`.

**E4.8 Анти-накрутка (FR-18).** Лимит 1 отзыв на пару (user,institution,child) на уровне БД; rate limit на создание; velocity-детектор → авто-флаг (`flagged_reason`) на приоритетную ручную модерацию, не автоблокировка.

**Критерии готовности:** тест формулы decay на фиксированных часах (`platform/clock`); тест «повторная доставка события не искажает агрегат»; тест «институция с 3 отзывами → `rating_status=insufficient`».

---

### ВЕХА 5 — Communications

**E5.1 Схема `communications`.** 6 таблиц; `applications` UNIQUE(`applicant_id`,`vacancy_id`), `employer_responses` UNIQUE(`institution_id`,`applicant_id`) — идемпотентность на уровне БД (фронт уже полагается на это, `web/lib/data.ts:2126-2127`).

**E5.2 WebSocket-хаб и мультиинстансность.** Соединения в памяти инстанса; кросс-инстансная доставка — Redis Pub/Sub (`chat:conv:{id}`). Сообщение сначала пишется в Postgres (источник истины), затем публикуется. Клиент дедуплицирует по `message_id`, история — REST-запросом, WS только за «живую» доставку. N инстансов без sticky-сессий.
Профилактика утечек: 2 горутины на соединение (reader/writer), снимаются по общему `CancelFunc`; `SetReadDeadline`+ping/pong 30с; исходящий канал с буфером 64, при переполнении — закрытие соединения, не блокировка хаба. Тесты — `-race` + тест возврата `NumGoroutine()` к базовому после закрытия 100 соединений.
Безопасность: токен — в `Sec-WebSocket-Protocol`, не в query (query попадает в логи прокси).

**E5.3 Уведомления.** Redis Streams + consumer group. Адаптеры провайдеров за интерфейсом `notify.Sender`; для MVP реальный — только in-app, SMS/email — заглушки с логированием (провайдер для РТ не выбран). Ретраи с backoff, DLQ-стрим после N попыток.

**E5.4 Вакансии и отклики.** CRUD вакансий (владелец институции), `POST /vacancies/{id}/apply` (идемпотентно), `POST /applicants/{id}/contact` (идемпотентно).

**E5.5 Профиль соискателя и видимость.** `visibility ∈ {draft,on_response,public}` + `hide_contacts`. Фильтрация контактов — на уровне SQL-проекции (не читать лишнее из БД), не только в DTO-маппинге — минимизация PII значит «не прочитать», а не «прочитать и не отдать».

**E5.6 Загрузка CV в S3.** Presigned PUT, allow-list `application/pdf`, лимит 5MB, ключ генерируется сервером.

**Критерии готовности:** интеграционный тест с двумя экземплярами хаба на одном Redis (сообщение из хаба A доходит до клиента хаба B); тест на отсутствие утечки горутин; тест «повторный apply не создаёт второй `application`».

---

### ВЕХА 6 — Analytics

**E6.1 Схема `analytics`.** `profile_events` — партиции по месяцам (`PARTITION BY RANGE(created_at)`), поля включают `ip_hash` (хешированный, не сырой IP — PII), `user_agent_class` (bot/mobile/desktop, не полная строка). `profile_events_daily` — агрегат с PK для идемпотентного upsert.

**E6.2 Путь записи не блокирует запрос.** Handler кладёт событие в bounded-канал (буфер 10k); batch-writer пишет `COPY FROM` каждые 200мс/500 событий. Переполнение канала → событие отбрасывается + инкремент `analytics_events_dropped_total` (аналитика — lossy by design, осознанный trade-off против p95 основного пути).

**E6.3 Ночная агрегация.** Джоб в `cmd/worker` под advisory-lock, идемпотентный `INSERT ... ON CONFLICT DO UPDATE` за прошедшие сутки + пересчёт «вчера» на поздние события.

**E6.4 Retention.** Сырые события — 90 дней, удаление через `DROP PARTITION` (мгновенно, без блокировок и раздувания WAL).

**E6.5 API дашборда.** `GET /api/v1/institutions/{id}/analytics?from&to&granularity` — только владелец/модератор. CSV — потоковый (`csv.Writer` прямо в `ResponseWriter`). **PDF-экспорт выносится в конец, делается только по отдельному подтверждению** (внешняя зависимость + шрифты под кириллицу и таджикские диакритики — отдельный риск).

**E6.6 Наблюдаемость (закрывающий эпик).** Prometheus RED-метрики, `/metrics`, алерт на p95>300мс и error-rate.

**Критерии готовности:** нагрузочный тест — включение аналитики не сдвигает p95 каталога; тест «переполнение канала не блокирует handler»; в `profile_events` нет сырых IP/телефонов (тест схемы).

---

## Сводная таблица вех

| # | Веха | Размер | Зависит от | Разблокирует | Главный технический риск |
|---|---|---|---|---|---|
| 0 | Платформенный скелет | M | — | всё | Выбор миграционного инструмента (goose — решено) |
| 1 | Catalog read-path | L | 0 | переезд фронта с mock; вехи 3,4 | p95 на гео+trgm; N+1 в карточке |
| 2 | Auth / RBAC | L | 0, 1 (FK на institutions) | вехи 3,4,5,6 | Refresh-reuse detection; PII-минимизация children |
| 3 | Catalog write + audit + идемпотентность | M | 1, 2 | веха 4 (модерация), 5 (владение вакансиями) | Audit вне транзакции = дыра в trail; partial-update nil vs null |
| 4 | Reviews / Ratings | XL | 1, 2, 3 | доверие к продукту, монетизация | Decay-пересчёт; дубли в агрегате при at-least-once |
| 5 | Communications | XL | 2, 3 | HR-модуль, чат | WS на N инстансов; утечки горутин; PII соискателя |
| 6 | Analytics | M | 1, 2, 3 | дашборд учреждения, Pro/Enterprise | Влияние записи событий на p95 основного пути; PDF-экспорт |

Критический путь: **0 → 1 → 2 → 3 → 4**. Вехи 5 и 6 после 3 могут идти параллельно с 4 при втором исполнителе (пересечение файлов минимально — `cmd/api/routes.go` и общие миграции, конфликт разрешается нумерацией миграций и отдельными файлами роутов).

---

## Risks

**1. `nil` deref / неправильная обработка указателей.** Наибольший риск в вехе 1 (десятки optional-полей `Institution`) и вехе 2 (`auth.FromContext`). Митигация: аксессоры контекста всегда `(T,bool)`; линтеры `nilnil/nilerr` с вехи 0; тест «NULL в БД → nil в домене → отсутствие ключа в JSON» на каждое optional-поле; `logx.PtrOrNil` — единственный способ логировать `*T`. Отдельная ловушка: `transport=false` vs `transport=nil` в фильтрах — все булевы фильтры `*bool`.

**2. Goroutine leaks / race conditions.** С вехи 4 (релей outbox), острее в вехе 5 (WS-хаб) и 6 (batch-writer). Митигация: `go test -race` в CI на каждый push; тест `NumGoroutine()` до/после жизненного цикла компонента; запрет «голых» `go func()` — только `errgroup.WithContext`+`SetLimit`; каждый долгоживущий компонент имеет `Close(ctx) error`, вызываемый из graceful shutdown.

**3. Неправильное размещение валидации (дубли/рассинхрон контрактов).** Митигация: правило зафиксировано в Architecture (нужно состояние → usecase, не нужно → transport); архитектурный тест `go list -deps ./internal/*/domain` не должен содержать `net/http`/`pgx`.

**4. Дрейф контракта с уже написанным фронтом.** Три известных расхождения:
- `Institution.id` — на фронте `number` (`web/lib/data.ts:192`), в БД `UUID`. Требует правки фронта (`instId` везде).
- `Institution.type` — на фронте одиночный `CategoryKey`, в БД `TEXT[]` (мульти-тип, SRS v2.1). API отдаёт `types: string[]`, фронт до адаптации берёт `types[0]`.
- `ChildLink.name` — фронт хранит имя ребёнка (`web/lib/app-state.tsx:23`), бэкенд принципиально не принимает (PII). Требует правки фронта — иначе POST падает на `DisallowUnknownFields`.
Митигация: зафиксировать OpenAPI-спеку эндпоинтов каталога как `backend/docs/openapi.yaml`, вести вместе с кодом.

**5. p95≤300мс не достигнут на реальных данных.** Сидовые 9 институций ничего не доказывают. Митигация: в задаче 31 сгенерировать 5-10 тыс. синтетических институций, зафиксировать baseline в `backend/docs/perf-baseline.md`. План Б при недостаточной скорости trgm-поиска: материализованный `tsvector`+`unaccent`.

**6. Redis как точка отказа.** Митигация: короткие таймауты + деградация везде (кэш промахнулся → БД; лимитер недоступен → fail-open с алертом, не fail-closed). Тест деградации — в задаче 31.

**7. Audit trail с дырами.** Митигация: `Recorder.Record` принимает `pgx.Tx`, не пул — сигнатура физически не даёт записать вне транзакции. Policy-тест из вехи 3 перебирает все мутирующие роуты.

**8. At-least-once доставка искажает агрегаты рейтинга.** Митигация: консьюмер идемпотентен по `event_id`; ночной полный пересчёт — самовосстановление от накопленных расхождений.

**9. Соблазн «дотянуть» вместо перепланирования.** Ожидаемые точки пересмотра: после задачи 24 (если один LATERAL-запрос медленнее раздельных — пересмотреть решение по параллелизму), после задачи 31 (если p95 не достигнут — менять стратегию поиска/кэша), перед вехой 4 (модель метрик может измениться после реальных данных).

---

## Out of scope

Явно НЕ входит в план, не реализуется без отдельного запроса:

- **Kubernetes**, Helm, service mesh (Docker для MVP по CLAUDE.md).
- **Elasticsearch** (сознательное решение — `pg_trgm`/`tsvector` покрывает MVP).
- **Микросервисная сетка** — один бинарник, границы логические.
- **Вторая волна по CLAUDE.md**: GPS-трекинг транспорта, онлайн-оплата (Корти Милли/Alif Pay — только описаны, без интеграции), нативные мобильные приложения, AI-подбор школы, CV-база с обратным поиском (HR-модуль), конструктор меню питания.
- **Отзывы соискателей о работодателе** (Glassdoor-стиль) — отложено ранее.
- **PDF-экспорт аналитики** — конец вехи 6, по отдельному подтверждению.
- **Антивирус-сканирование загружаемых файлов** — известный пробел вех 3/5, принятый риск MVP.
- **Реальные SMS/email-провайдеры** — заглушки, провайдер для РТ не выбран.
- **Правки фронта `web/`** — план описывает только бэкенд. Три контрактных расхождения (см. Risks п.4) требуют отдельной задачи по фронту.
- **Миграция продовых данных** — продовых данных нет, только dev-сид.
- **Формальный путь апелляции при отказе в регистрации учреждения** — сознательно не реализуется в MVP (см. раздел «Модерация»).

---

## Design decisions log (design approval gate — пройден)

| # | Решение | Кем принято |
|---|---|---|
| 1 | 8 метрик рейтинга (не 9) | Определено по `EduHub_Functional_Requirements.md` v1.1 в разговоре ранее |
| 2 | Миграционный инструмент — `pressly/goose/v3` | Вынуждено правилами проекта (`CONCURRENTLY` требует `NO TRANSACTION`) |
| 3 | Module path — `github.com/abdulhalim/eduhub/backend` | Дефолт, не критично |
| 4 | `moderation.audit_log` — отдельная схема `moderation` | Пользователь подтвердил |
| 5 | `auth.children.institution_id` — обычный физический FK (`ON DELETE RESTRICT`), не eventual-consistency порт | Пользователь подтвердил (изменяет рекомендацию go-planner) |
| 6 | Верификация владельца учреждения — один модератор, не два | Пользователь выбрал MVP-вариант |
| 7 | Отказ в регистрации учреждения — без формального пути апелляции | Пользователь выбрал MVP-вариант |
| 8 | Очередь модерации — одна общая, без деления по региону | Пользователь выбрал MVP-вариант |
| 9 | Claim-паттерн, структурированная причина отказа, приоритет по флагам, уведомление актора, self-moderation guard | Приняты как общие практики по умолчанию, не оспорены |

**Релевантные файлы фронта для сверки контрактов:** `web/lib/data.ts` (`Institution:191`, `Review:1184`, `Vacancy:1611`, `Applicant:1993`, `Application:2128`, `EmployerResponse:2112`, метрики `822-829`), `web/lib/app-state.tsx` (`ChildLink:21`, `RegisterInstitutionInput:31`), `web/app/(site)/search/page.tsx:56-116` (фильтры каталога), `web/lib/geo.ts` (клиентское определение координат).
