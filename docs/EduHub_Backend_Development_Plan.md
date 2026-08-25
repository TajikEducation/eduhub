# EduHub — план разработки бэкенда (v1)

**Дата:** 2026-08-24, ревизия 2026-08-25 (go-reviewer аудит — 10 критичных находок закрыты, см. Design decisions log)
**Статус:** согласован (design approval gate пройден), готов к реализации
**Авторы:** go-advisor (архитектурная консультация) + go-planner (детальный план) + go-reviewer (аудит) + согласование с пользователем

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
- **Минимизация PII детей**: `auth.children` хранит только `age_group` + `status` + `institution_id` (+ `confirmation_status` — см. ниже). Имя/фото ребёнка бэкенд не принимает, не хранит, не логирует — ломает текущий фронтовый `ChildLink.name`, правка фронта нужна отдельно (см. Risks).
- **Право на удаление аккаунта (закон РТ №1537)**: soft-delete + анонимизация, не физический `DELETE` — см. `docs/EduHub_Database_Schema.md`, раздел `auth.users`.
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

**Rate-limiting — покрытие по эндпоинтам (добавлено 2026-08-25).** Механизм спроектирован один раз в вехе 0 (задача 15, in-memory token bucket по IP, Redis-based версия — веха 5 при мультиинстансе), но применение к эндпоинтам сверх `/auth/login` не было зафиксировано явно. Список обязательных точек применения:

| Эндпоинт | Ключ лимита | Обоснование |
|---|---|---|
| `/auth/login` | email+IP | anti-bruteforce (E2.7), уже в плане |
| `/auth/verify/resend` | email+IP | FR-40 — защита от накрутки почтового провайдера |
| `POST` чат-сообщение (`communications.messages`) | `sender_id` | защита от спама в переписку (E5.2) |
| `POST` создание нового диалога (`communications.conversations`) | `user_id` за период | отдельно от лимита сообщений — защита от «написать сразу многим одно и то же» (не только учреждениям — после обобщения на родитель↔родитель тот же вектор) |
| `POST` загрузка медиа (`institutions/{id}/media`, CV) | `institution_id`/`user_id` | защита воркера сжатия от долбёжки запросами (E3.6, E5.6) |
| `POST` создание отзыва | `user_id`+IP | FR-18, привязка к конкретному механизму (E4.8 упоминал словами, без реализации) |
| `POST` создание отзыва о работодателе | `user_id`+IP | тот же принцип, что родительский отзыв (E4.9) |
| `POST /communications/visit_requests` | IP+телефон | публичный эндпоинт, собирает PII от гостя без аутентификации |
| `POST /communications/employer_responses` | `institution_id` за период | FR-37 — не разовый rate limit, а именно лимит за период (см. находку про пожизненную уникальность, остаётся отдельным пунктом) |

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

**2. `docker-compose.yml` — Postgres+PostGIS, Redis, MinIO**
Что: `postgis/postgis:16-3.4` (порт 5433), `redis:7-alpine`, `minio/minio` (добавлено 2026-08-25 — S3-совместимое хранилище, **только dev/test**, в проде реальный S3-совместимый провайдер по конфигу; нужен уже с вехи 3, где загрузка медиа переведена на multipart через бэкенд — без него интеграционные тесты загрузки не на чем гонять), healthcheck, init-скрипт создаёт `eduhub_test`.
Приёмка: `docker compose up -d && docker compose ps` — все три `healthy`; `psql "$DATABASE_URL" -c "SELECT postgis_version()"` печатает версию; `curl -f http://localhost:9000/minio/health/live` → 200.

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

**16. CI, линтеры, Dockerfile**
Что: `.golangci.yml` (`errcheck/govet/staticcheck/nilerr/nilnil/bodyclose/rowserrcheck/contextcheck/exhaustive/gosec/noctx`), GitHub Actions job `backend`: `go vet` → `golangci-lint run` → `go test -race ./...` → `go test -tags=integration ./...` (сервисный контейнер postgis) → `govulncheck ./...` → **`docker build`** (добавлено 2026-08-25). Deploy-артефакт: `Dockerfile` — multi-stage build (`.claude/rules/devops.md`), builder-слой `golang:1.23` собирает статический бинарник (`CGO_ENABLED=0`), финальный слой `distroless/static` (без шелла/пакетного менеджера — меньше поверхность атаки), `.dockerignore` исключает `.git`/тесты/локальные `.env`, секреты — только через переменные окружения контейнера, не в образе. Без CI-шага сборки образа можно накопить дрейф между локальной сборкой и тем, что реально задеплоится.
Приёмка: `make lint` → 0 issues; `make vuln` → No vulnerabilities; `docker build -t eduhub-api .` без ошибок, `docker run --rm eduhub-api` (с валидным `DATABASE_URL` и т.д.) поднимает сервер, `/healthz` отвечает 200 изнутри контейнера; push в ветку → CI job зелёный, включая шаг `docker build`.

**Критерий готовности вехи 0:** `make test-race && make lint && make vuln` зелёные; `make migrate-up && make migrate-down` идемпотентны; `/healthz` 200, `/readyz` 503 при остановленном Postgres; `docker build` проходит в CI.

---

### ВЕХА 1 — Catalog read-path (задачи 17-31, полная детализация)

Цель: публичный API каталога, на который фронт (`web/app/(site)/search/page.tsx`, `web/components/InstitCard.tsx`, `web/app/(site)/institutions/[id]/page.tsx`) переключается с mock-данных.

Контракт эндпоинтов:
- `GET /api/v1/institutions?q&type&region&area&min_price&max_price&min_rating&transport&food&verified&curriculum&program_level&discount&sort&lat&lng&radius_km&limit&cursor` (`curriculum`/`program_level` добавлены 2026-08-25 — были в FR-01, индексы уже в задаче 18, но не попали в исходный контракт; `discount` — тоже прямое требование FR-01, ранее пропущен)
- `GET /api/v1/institutions/{id}`
- `GET /api/v1/institutions/{id}/news?limit&cursor`
- `GET /api/v1/reference/regions`

Фильтры — 1:1 с реальным фронтом (`web/app/(site)/search/page.tsx:56-116`), плюс гео (`web/lib/geo.ts` уже умеет определять координаты клиентски).

**17. Миграция 00002 — `catalog.institutions`**
Поля: `id UUID PK DEFAULT gen_random_uuid()`, `name JSONB NOT NULL`, `types TEXT[] NOT NULL`, `region TEXT NOT NULL`, `city JSONB`, `district TEXT`, `description JSONB`, `address JSONB`, `geo GEOGRAPHY(Point,4326) NOT NULL`, `license_no TEXT`, `languages TEXT[]`, `program_level TEXT[]`, `curriculum TEXT[]`, `price INT`, транспорт/питание/скидки-поля, `phone TEXT`, `email TEXT`, `website TEXT`, `socials JSONB`, `cover_photo_s3_key TEXT`, `age_range TEXT`, `tag JSONB` (полный список — см. `docs/EduHub_Database_Schema.md`, обновлено 2026-08-25 после аудита: изначальный список не покрывал контакты/описание, требуемые FR-06), `verified BOOL NOT NULL DEFAULT false`, `moderation_status TEXT NOT NULL DEFAULT 'pending' CHECK (moderation_status IN ('pending','approved','rejected'))`, `plan TEXT NOT NULL DEFAULT 'free'`, `founded INT`, `students_count INT`, `rating_avg NUMERIC(3,2)`, `review_count INT NOT NULL DEFAULT 0` (денормализация — заполняется вехой 4 через порт `RatingSync`, но колонки нужны сразу — фильтр `min_rating` и `sort=score` обязаны работать одним SQL-запросом), `created_at/updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`.
Приёмка: `make migrate-up` → OK; `\d catalog.institutions` показывает все колонки; `make migrate-down` откатывает чисто.

**18. Миграция 00003 — индексы `CONCURRENTLY`**
`-- +goose NO TRANSACTION`. `GIST(geo)`; `GIN(name jsonb_path_ops)`; `GIN((name->>'ru') gin_trgm_ops)` и аналогично `'tg'` (подстрочный поиск на обоих языках); `GIN(types)`; `GIN(curriculum)`; `GIN(program_level)`; `btree(region, district)`; частичный `btree(rating_avg DESC, id) WHERE moderation_status='approved'`; `btree(price) WHERE moderation_status='approved'`; частичный `UNIQUE(lower(name->>'ru'), region, district) WHERE moderation_status<>'rejected'` (не через `CONCURRENTLY` — уникальные constraint'ы создаются обычным `CREATE UNIQUE INDEX CONCURRENTLY` тоже допустимо вне транзакции, тот же `NO TRANSACTION` файл).
Решение по поиску: `pg_trgm`, не `tsvector` — словаря для таджикского в Postgres нет, `to_tsvector('simple',...)` не даёт стемминга ни для одного языка, а фронт сейчас делает подстрочное совпадение (`.includes`) — trigram воспроизводит это поведение на обоих языках одинаково.
Приёмка: `make migrate-up` без ошибок; `EXPLAIN` на гео-запросе показывает `Index Scan ... gist`.

**19. Миграция 00004 — сателлитные таблицы каталога**
`catalog.institution_staff`, `catalog.achievements` (полиморфно: `owner_type CHECK IN ('institution','staff','student')` + `owner_id UUID`, индекс `(owner_type, owner_id)`, без FK — полиморфные ссылки не поддерживают FK), `catalog.institution_gallery` (S3-ключ, не URL), `catalog.institution_alumni`, `catalog.news_articles`, `catalog.institution_metrics` (пустая на этой вехе — заполняется портом `RatingSync` начиная с вехи 4, структура создаётся сразу вместе с остальным каталогом). Все — FK на `catalog.institutions(id) ON DELETE CASCADE` (внутри своей схемы FK разрешены свободно).
Приёмка: `make migrate-up/down` чисто; 7 таблиц в схеме `catalog`.

**20. `internal/catalog/domain` — сущности и `Filter`**
RED: (а) `Filter{MinPrice:p(500),MaxPrice:p(100)}.Validate()` → ошибка поля `min_price`; (б) `Limit=0` после `Normalize()` → 20, `Limit=500` → капается на 50; (в) `Lat` без `Lng` → ошибка; (г) `RadiusKm` без координат → ошибка; (д) пустой `Filter` валиден. Конструктор `Institution` гарантирует `Gallery`/`Types` как `[]T{}`, не `nil`. Поля `Curriculum []string`/`ProgramLevel []string`/`Discount *bool` (добавлено 2026-08-25) — множественные, пересечение (не точное совпадение).
Приёмка: `go test ./internal/catalog/domain/ -v` → PASS на 6 кейсах; `go list -deps ./internal/catalog/domain | grep -E 'pgx|net/http'` → пусто.

**21. `internal/catalog/repo/postgres` — `List` без гео**
RED (integration, транзакция+rollback фикстур): 5 институций (3 approved, 1 pending, 1 rejected). (а) пустой фильтр → только 3 approved; (б) `Region="sughd"` → 1; (в) `Types=["cat_school"]` → верное подмножество; (г) `Q="гулис"` находит по подстроке в обоих языках; (д) `MinRating=4.5` не возвращает институции с `rating_avg IS NULL` (NULL — «нет данных», не «ноль»); (е) `Curriculum=["bilingual","stem"]` возвращает институции хотя бы с одним из значений (SQL `&&`, оператор пересечения — FR-01 буквально «пересечение массивов», не `@>`/«содержит все»: родитель отмечает несколько программ и хочет видеть учреждения с любой из них, не только с обеими сразу); (ж) `Discount=true` → только `discount_available=true`.
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
RED (~14 кейсов, +2 добавлены 2026-08-25): `sort=hack` → 400 c полем `sort`; `min_price=abc` → 400; `lat=200` → 400; `limit=-1` → 400; `type=cat_school,cat_uni` → слайс из 2; `transport=1` → `true`, `transport=` → `nil` (не `false` — отсутствие фильтра и «выключен» это разное); `curriculum=bilingual,stem` → слайс из 2; `discount=true` → `*bool` true; неизвестный query-параметр игнорируется (форвард-совместимость с фронтом).
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

**E2.1 Схема `auth`.** `users` (`email` уникальный основной канал, `phone` nullable опциональный второй очереди, `password_hash` nullable — пусто для чисто-Google-аккаунтов, `role`, `status ∈ {unverified,active,banned,deleted}`, `email_verified_at`, `consent_at`+`consent_version`, `failed_login_count`, `locked_until`, `deleted_at`), `oauth_identities` (`provider='google'`, `provider_user_id`, UNIQUE пара), `verification_codes` (`channel`, `purpose`, `code_hash`, `attempts_count`, `expires_at`), `refresh_tokens` (`token_hash`, `family_id`, `expires_at`, `revoked_at`, `replaced_by`), `children` (`user_id`, `institution_id` **FK на `catalog.institutions` ON DELETE RESTRICT** — см. Architecture, `age_group`, `status ∈ {current,alumnus,transferred}`, `confirmation_status ∈ {pending,confirmed,rejected}` + `confirmed_by`/`confirmed_at`, БЕЗ имени/фото, `UNIQUE(user_id,institution_id)`).

**Приоритет каналов регистрации (пересмотрено 2026-08-25):** email/пароль + вход через Google — основные пути. Телефон — опциональное поле, SMS-верификация и связанный с ней rate-limit по номеру — вне MVP (SMS-провайдер в РТ платный за отправку, не оправдан на холодном старте). Отменяет исходную формулировку FR-40 (SRS v2.1, «телефон — основной канал»); `docs/EduHub_Functional_Requirements.md` синхронизирован в этой ревизии, SRS `.docx` — отложено по решению пользователя.

**E2.2 Хеширование и парольная политика.** argon2id, параметры в конфиге, PHC-формат хранения (параметры можно поднять без слома существующих хешей). Google-only аккаунты (`password_hash IS NULL`) — вход только через `oauth_identities`, `/auth/login` для них возвращает `google_account_no_password`.

**E2.3 JWT и ротация refresh.** Access 15 мин, refresh 30 дней, хранится только хеш. **Reuse detection**: предъявление уже использованного refresh отзывает всю `family_id` + запись в audit — с самого начала, не «потом». HS256 достаточен пока валидатор один (монолит); переход на RS256/JWKS — если появится второй потребитель токенов.

**E2.4 Эндпоинты.** `POST /auth/register` (email+пароль; идемпотентен через `platform.idempotency_keys` с nullable `user_id` и статусом `in_progress`/`completed` — см. схему; повтор с тем же email → 409 `email_taken`), `POST /auth/oauth/google` (обмен Google id-token → сессия, создаёт `users`+`oauth_identities` при первом входе), `POST /auth/verify` (код подтверждения email), `POST /auth/verify/resend` (rate-limit по email+IP), `POST /auth/password/reset-request`, `POST /auth/password/reset-confirm`, `POST /auth/login` (единое сообщение об ошибке для «нет юзера» и «неверный пароль» — anti-enumeration), `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`, `POST /auth/consent` (фиксация `consent_at`+`consent_version` — обязательный шаг регистрации, не мелкий текст внизу формы), `DELETE /auth/me` (soft-delete + анонимизация — закон РТ №1537, см. схему `auth.users`).

**E2.5 Middleware RBAC.** `RequireAuth` кладёт `auth.Principal{UserID,Role}` в контекст; аксессор `FromContext(ctx)(Principal,bool)` — обязательно с `ok`. `RequireRole(roles...)`. Табличный тест матрицы «роль × эндпоинт» — растёт в вехах 3-6.

**E2.6 Children CRUD с PII-минимизацией и подтверждением учреждением.** API принимает только `institution_id`, `age_group`, `status`. Явный тест: поле `name` в запросе → 400 `unknown field`. **Проверка статуса институции — НЕ FK, а явный guard в usecase**: FK на `catalog.institutions` гарантирует только существование строки (ссылочную целостность), не значение `moderation_status` — привязка к `pending`/`rejected` учреждению технически пройдёт FK, поэтому usecase обязан отдельно проверить `moderation_status='approved'` через порт `catalog.InstitutionStatus` до вставки. Новое: привязка создаётся со `confirmation_status='pending'`; отдельный эндпоинт в кабинете учреждения — `GET /institutions/{id}/children/pending` (роль `institution`, свой институт) + `POST /children/{id}/confirm|reject` (структурированная причина при reject, запись в `moderation.audit_log`). Право на отзыв (веха 4, E4.2) требует `confirmed`, не просто существования строки.

**E2.7 Anti-bruteforce.** Rate limit на `/auth/login` по email и IP, экспоненциальная задержка, блокировка на 15 мин после N попыток; события в audit без пароля, email маскируется в логах (например `a***@example.com`).

**Критерии готовности:** матричный тест ролей зелёный; тест reuse-detection refresh-токена зелёный; тест «FK не заменяет проверку approved-статуса» (привязка к `pending`-институции должна быть отклонена usecase, не только полагаться на FK); ни один тест не пишет PII в лог (grep-проверка тестового вывода); `-race` чистый.

---

### ВЕХА 3 — Catalog write + RBAC-защита мутаций + сквозной audit

**E3.1 Пакет `internal/moderation`.** Схема `moderation.audit_log` (`actor_id`, `actor_role`, `action`, `target_type`, `target_id`, `reason`, `payload_diff JSONB`, `request_id`, `created_at`). Порт `moderation.Recorder.Record(ctx, tx, Entry) error` — принимает транзакцию (запись audit физически невозможна вне транзакции изменения). Здесь же — очередь модерации с claim-паттерном (`claimed_by`, `claimed_at`, `priority`, `flagged_reason`) — см. раздел «Модерация — детальный дизайн» выше.

**E3.2 Пакет `internal/platform/idempotency`.** `platform.idempotency_keys` (`key`,`endpoint`,`user_id` nullable,`request_hash`,`status ∈ {in_progress,completed}`,`response_status`,`response_body`,`created_at`, UNIQUE(`key`,`endpoint`)). Строка вставляется со статусом `in_progress` ДО выполнения операции — атомарный INSERT работает как распределённая блокировка против гонки параллельных повторов (два одновременных ретрая не оба проходят проверку). Тот же ключ+тело, статус `completed` → сохранённый ответ; тот же ключ+другое тело → 422 `idempotency_key_reuse`; тот же ключ, статус всё ещё `in_progress` → 409 (уже обрабатывается). TTL 24ч. `user_id` nullable — нужен для анонимных мутирующих эндпоинтов (`/auth/register`). Переиспользуется в вехах 4-5.

**E3.3 Регистрация институции.** `POST /api/v1/institutions` (роль `user`/`institution`), `moderation_status='pending'`, владелец в `catalog.institution_owners`. Идемпотентность — два независимых механизма: (1) `Idempotency-Key` защищает от повтора одного и того же HTTP-запроса (сетевой ретрай); (2) частичный `UNIQUE(lower(name->>'ru'), region, district) WHERE moderation_status<>'rejected'` в схеме защищает от осознанного двойного сабмита формы без заголовка (дубль-клик) — конфликт возвращает `institution_already_registered`. Форма — минимум из `RegisterInstitutionInput` (`web/lib/app-state.tsx:31+`), остальное — дефолты.

**E3.4 Редактирование профиля.** `PATCH /api/v1/institutions/{id}` — владелец или модератор. **Оптимистическая блокировка — только через ETag/`updated_at` (уточнено 2026-08-25, `version`-поле не заводим)**: клиент обязан прислать `If-Match` со значением ETag, полученным на последнем `GET` (тот же ETag из задачи 28, вычисляется из `updated_at`+id). Handler пересчитывает текущий ETag из актуального `updated_at`, сверяет с `If-Match` — не совпадает → `412 Precondition Failed`, конкурентная правка не «последний победил». Один механизм на чтение (`If-None-Match`/304) и запись (`If-Match`/412) — не два несинхронизированных (изначально план упоминал ETag и отдельный `version INT` как будто взаимозаменяемые, что требовало бы клиенту знать оба значения раздельно). Частичное обновление: `nil` = не трогать, явный JSON `null` = очистить (нужен custom-unmarshal, отдельная задача при реализации).

**E3.5 Модерация.** `POST /api/v1/moderation/institutions/{id}/approve|reject` (роль `moderator`/`admin`, обязательный структурированный `reason` для reject — без апелляции, см. раздел выше), `GET /api/v1/moderation/queue` (одна общая очередь, claim-эндпоинт). Каждое действие — транзакция {смена статуса + audit + инвалидация кэша + уведомление актору}. Отдельно (добавлено 2026-08-25): `POST /api/v1/institutions/{id}/owner-verification` (учреждение подаёт документ или запрашивает `manual_exception`) + `POST /api/v1/moderation/owner-verifications/{id}/approve|reject` (модератор) — материализует FR-34 в `catalog.institution_owner_verifications` (раньше был только булев результат `verified`, без хранимого документа/обоснования). Для гос.школ/садов, освобождённых от лицензирования по закону РТ, и для мелких центров без формальных документов — `document_type='manual_exception'` с обязательным `verification_notes`.

**E3.6 Медиа (пересмотрено 2026-08-25 — было presigned PUT, отказались).** `POST /api/v1/institutions/{id}/media` — **multipart-загрузка через бэкенд**, не presigned PUT в S3. Причина отказа от presigned: (1) FR-07 требует серверное автосжатие фото до ~200КБ — при прямой загрузке в S3 бэкенд не видит байты, сжимать нечего; (2) без подтверждающего шага после presigned-загрузки остаются «осиротевшие» ссылки в БД, если клиент запросил URL, но так и не залил файл (закрыл вкладку/потерял сеть) — с multipart записи в БД просто не бывает без успешной загрузки. Поток: `multipart-запрос → лимит размера+allow-list content-type в handler → воркер сжатия (библиотека обработки изображений) → запись файла в S3 → запись метаданных в `institution_gallery`/`applicants.cv_s3_key` в одной транзакции`. Лимит числа фото на профиль — по тарифу (см. `docs/EduHub_Pricing_Tiers.md`), проверяется в usecase до принятия файла. Имя объекта генерируется на сервере (защита path traversal). Антивирус-сканирование — вне MVP, известный пробел.

**E3.7 Инвалидация кэша.** Запись инкрементирует `catalog:version` в Redis — после approve список отдаёт новую институцию без ожидания TTL.

**Критерии готовности:** ни один мутирующий эндпоинт не проходит тест без записи в `moderation.audit_log` (policy-тест по всем зарегистрированным мутирующим роутам); повторный POST с тем же `Idempotency-Key` не создаёт дубль.

---

### ВЕХА 4 — Reviews / Ratings

**E4.1 Схема `reviews`.** `reviews` (UNIQUE(`user_id`,`institution_id`,`child_id`) — идемпотентность на уровне БД), `review_metrics` (нормализовано по строкам: `review_id`,`metric_key`,`score 1-5`, UNIQUE(`review_id`,`metric_key`)), `institution_rating_agg` (`institution_id`,`metric_key`,`weighted_avg`,`review_count`,`updated_at`). Плюс (добавлено 2026-08-25, см. E4.9) `employer_reviews`+`employer_review_metrics` — полностью отдельный контур для отзывов о работодателе.
**8 метрик** (решено ранее в разговоре — по `docs/EduHub_Functional_Requirements.md` v1.1, уточнённая версия; девятая величина в SRS/бизнес-плане — вычисляемое среднее, не отдельная оцениваемая метрика): `quality, conditions, safety, food, transport, price, parent_involvement, inclusivity`.

**E4.2 Eligibility (верификация отзыва).** Разрешено только при наличии `auth.children` со связью user→institution в статусе `current`/`alumnus`/`transferred` **И `confirmation_status='confirmed'`** (учреждение подтвердило привязку — см. E2.6, добавлено 2026-08-25: самодекларации родителя одной больше недостаточно, это вторая линия защиты сверх `UNIQUE(user_id,institution_id)`). Проверка — в usecase через порт `auth.ChildLinkExists` (кросс-схемная логика — не через FK, поскольку `reviews` не владеет схемой `auth`, порт остаётся правильным паттерном в этом направлении, в отличие от `children→institutions`).

**E4.3 Recency decay.** `w_i = 0.5 ^ (age_days / H)`, `H=365` дней (конфиг, не хардкод). `weighted_avg = Σ(w_i·score_i) / Σ(w_i)` по метрике. Порог публикации: `review_count >= MIN_REVIEWS` (старт 5), иначе `rating: null` + `rating_status: "insufficient"`. Два триггера пересчёта: (1) точечный по событию «отзыв одобрен/отклонён», (2) обязательный ночной полный пересчёт (веса меняются от одного лишь течения времени).

**E4.4 Outbox + релей.** `reviews.outbox` (`id`,`topic`,`payload`,`created_at`,`published_at`). Публикация отзыва = {отзыв+метрики+outbox-запись} в одной транзакции. Релей — `FOR UPDATE SKIP LOCKED` батчами → Redis Stream → консьюмер идемпотентен по `event_id`.

**E4.5 Синхронизация с каталогом.** Консьюмер вызывает порт `catalog.RatingSync.Apply(ctx, instID, metrics []MetricScore)` — **передаёт массив всех 8 метрик**, не только общее среднее (исправлено 2026-08-25: карточка учреждения рендерит разбивку по 8 метрикам, одного `avg` недостаточно для собственного одного-запросного чтения карточки; `catalog` не вправе читать схему `reviews` напрямую). `catalog` сам пишет в свои денормализованные колонки — `catalog.institutions.rating_avg`/`review_count` (общее) + `catalog.institution_metrics` (по каждой из 8 метрик, новая сателлит-таблица) — бампает cache version. Владение схемами соблюдено.

**E4.6 Фоновые задачи и лидерство.** Ночной пересчёт и чистка idempotency-ключей — в `cmd/worker`, лидер через `pg_try_advisory_lock` — безопасно на нескольких инстансах без внешнего оркестратора.

**E4.7 Модерация отзывов и споры.** Статусы `pending/approved/rejected/disputed→resolved_kept|resolved_removed` (обновлено 2026-08-25 — `disputed` не конечное состояние), ответ учреждения (`reply`), формальный спор с SLA 72ч (`dispute_deadline`, `disputed_by`, `dispute_reason`, джоб-эскалация) — здесь путь апелляции ЕСТЬ (в отличие от отказа в регистрации учреждения, см. раздел «Модерация» выше). **Критерий разрешения спора сверяется со снимком, не с текущим состоянием**: при публикации отзыва (E4.2) в строку копируется `verified_at_publish` — было ли `confirmation_status='confirmed'` в тот момент; иначе разрешение спора зависело бы от состояния `auth.children`, которое могло измениться уже после публикации (родитель удалил привязку, учреждение отозвало подтверждение), что несправедливо по отношению к обеим сторонам. Каждое действие — в `moderation.audit_log`.

**E4.9 Отзывы об учреждении как о работодателе (Glassdoor-модуль, добавлено 2026-08-25, ранее отложено в 7-й волне фронта).** Полностью отдельный контур от родительских отзывов — своя верификация, свои метрики, ограниченная видимость:
- **Верификация трудоустройства — модератором платформы, НЕ учреждением.** `auth.employment_claims` (сотрудник заявляет `current`/`former` + опциональный документ-доказательство), решение — `pending`/`verified`/`rejected` через обычную очередь модерации. Если бы подтверждало само учреждение (как `auth.children`), оно могло бы просто не подтверждать авторов негативных отзывов — это убивает цель честного отзыва о работодателе.
- **5 метрик** (нормализовано по строкам, как `review_metrics`): `salary_conditions`, `management`, `team_atmosphere`, `workload`, `professional_growth`. Не влияют на `catalog.institution_metrics`/`rating_avg` — полностью изолированная агрегация, если она вообще понадобится (на MVP можно без агрегата, только список отзывов).
- **Модерация** — тот же цикл `pending→approved/rejected`, что и родительские отзывы (FR-16), ответ учреждения разрешён (не может скрыть отзыв, но может публично ответить).
- **Видимость жёстко ограничена**: `GET /api/v1/institutions/{id}/employer-reviews` требует авторизации И наличия у пользователя записи `communications.applicants` (то есть он соискатель). Родители/гости/обычные пользователи без анкеты соискателя не видят ни на публичном профиле учреждения, ни по прямой ссылке — это именно то, что снимает риск для монетизации: платящее учреждение никогда не видит эти отзывы на своей родительской витрине.
- Идемпотентность — `UNIQUE(user_id, institution_id)` на `employer_reviews`, один отзыв о работодателе на пару.

**E4.8 Анти-накрутка (FR-18).** Лимит 1 отзыв на пару (user,institution,child) на уровне БД; rate limit на создание; velocity-детектор → авто-флаг (`flagged_reason`) на приоритетную ручную модерацию, не автоблокировка.

**Критерии готовности:** тест формулы decay на фиксированных часах (`platform/clock`); тест «повторная доставка события не искажает агрегат»; тест «институция с 3 отзывами → `rating_status=insufficient`».

---

### ВЕХА 5 — Communications

**E5.1 Схема `communications`.** 8 таблиц (обновлено 2026-08-25 — было 6): `conversations` (один диалог на пару user/institution, `id` = `chat:conv:{id}`, раздельные `user_last_read_at`/`institution_last_read_at`) + `messages` (переструктурировано: `conversation_id FK`, `sender_type ∈ {user,institution}`, `sender_id` — раньше `messages` смешивала обе стороны в одной строке без различения отправителя, что делало чат нереализуемым); `notifications`; `vacancies`; `applicants`; `applications` UNIQUE(`applicant_id`,`vacancy_id`); `employer_responses` UNIQUE(`institution_id`,`applicant_id`) — идемпотентность на уровне БД (фронт уже полагается на это, `web/lib/data.ts:2126-2127`); `visit_requests` (новое — уже есть форма на фронте, `web/app/(site)/institutions/[id]/page.tsx:344-352`, но не было ни таблицы, ни эндпоинта; идемпотентно через частичный `UNIQUE(institution_id,phone) WHERE status='new'`, обязателен rate-limit по IP+телефону — публичный эндпоинт для гостя собирает PII).

**E5.2 WebSocket-хаб и мультиинстансность.** Соединения в памяти инстанса; кросс-инстансная доставка — Redis Pub/Sub (`chat:conv:{id}`, теперь backed реальной сущностью `communications.conversations.id`, не суррогатной парой). Сообщение сначала пишется в Postgres (источник истины), затем публикуется. Клиент дедуплицирует по `message_id` (= `messages.id`), история — REST-запросом, WS только за «живую» доставку. N инстансов без sticky-сессий. Rate-limit на отправку сообщения — по `sender_id`, защита от спама в чат. **Диалог обобщён 2026-08-25**: не только родитель↔учреждение, но и родитель↔родитель (`conversations.participant_a/b_type ∈ {user,institution}`, полиморфно). Отдельный rate-limit на **создание нового диалога** (не на сообщение внутри уже существующего) — за период на пользователя, применяется одинаково независимо от типа собеседника (тот же спам-вектор — написать сразу многим одно и то же). Механизм обнаружения (как пользователь находит `user_id` другого родителя, чтобы начать диалог) — открытый вопрос, не блокирует схему, решается отдельно до реализации.
Профилактика утечек: 2 горутины на соединение (reader/writer), снимаются по общему `CancelFunc`; `SetReadDeadline`+ping/pong 30с; исходящий канал с буфером 64, при переполнении — закрытие соединения, не блокировка хаба. Тесты — `-race` + тест возврата `NumGoroutine()` к базовому после закрытия 100 соединений.
Безопасность: токен — в `Sec-WebSocket-Protocol`, не в query (query попадает в логи прокси).

**E5.3 Уведомления.** Redis Streams + consumer group. Адаптеры провайдеров за интерфейсом `notify.Sender`. **Пересмотрено 2026-08-25**: push в браузере/PWA — реальный канал в MVP, не заглушка (Web Push API — встроенный браузерный стандарт через VAPID-ключи, бесплатно, не требует внешнего платного провайдера в отличие от SMS; NFR CLAUDE.md называет push основным каналом PWA — откладывать вместе с SMS/email было ошибкой плана, не осознанным решением). SMS/email остаются заглушками с логированием (провайдер для РТ не выбран — здесь задержка обоснована). `communications.push_subscriptions` хранит Web Push подписки (endpoint+ключи), `POST/DELETE /api/v1/push/subscribe`, `auth.users.notification_channels` — переключатели канала по пользователю. Ретраи с backoff, DLQ-стрим после N попыток.

**E5.4 Вакансии и отклики.** CRUD вакансий (владелец институции), `POST /vacancies/{id}/apply` (идемпотентно), `POST /applicants/{id}/contact` (идемпотентно).

**E5.5 Профиль соискателя и видимость.** `visibility ∈ {draft,on_response,public}` + `hide_contacts`. Фильтрация контактов — на уровне SQL-проекции (не читать лишнее из БД), не только в DTO-маппинге — минимизация PII значит «не прочитать», а не «прочитать и не отдать».

**E5.6 Загрузка CV в S3.** Multipart через бэкенд (тот же паттерн, что E3.6 — не presigned PUT), allow-list `application/pdf`, лимит 5MB, ключ генерируется сервером, запись `applicants.cv_s3_key` только после успешной загрузки.

**Критерии готовности:** интеграционный тест с двумя экземплярами хаба на одном Redis (сообщение из хаба A доходит до клиента хаба B); тест на отсутствие утечки горутин; тест «повторный apply не создаёт второй `application`».

---

### ВЕХА 6 — Analytics

**E6.1 Схема `analytics`.** `profile_events` — партиции по месяцам (`PARTITION BY RANGE(created_at)`), поля включают `ip_hash` (хешированный, не сырой IP — PII), `user_agent_class` (bot/mobile/desktop, не полная строка), `referrer_class` (добавлено 2026-08-25, FR-22/23: `search`/`catalog`/`map`/`direct`/`external` — классификация `Referer`-заголовка на handler'е, сырой URL не сохраняется — может нести PII в query источника). `profile_events_daily` — агрегат с PK для идемпотентного upsert.

**E6.2 Путь записи не блокирует запрос.** Handler кладёт событие в bounded-канал (буфер 10k); batch-writer пишет `COPY FROM` каждые 200мс/500 событий. Переполнение канала → событие отбрасывается + инкремент `analytics_events_dropped_total` (аналитика — lossy by design, осознанный trade-off против p95 основного пути).

**E6.3 Ночная агрегация.** Джоб в `cmd/worker` под advisory-lock, идемпотентный `INSERT ... ON CONFLICT DO UPDATE` за прошедшие сутки + пересчёт «вчера» на поздние события.

**E6.4 Retention.** Сырые события — 90 дней, удаление через `DROP PARTITION` (мгновенно, без блокировок и раздувания WAL).

**E6.5 API дашборда.** `GET /api/v1/institutions/{id}/analytics?from&to&granularity` — только владелец/модератор. CSV — потоковый (`csv.Writer` прямо в `ResponseWriter`). **PDF-экспорт выносится в конец, делается только по отдельному подтверждению** (внешняя зависимость + шрифты под кириллицу и таджикские диакритики — отдельный риск).

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
- **PDF-экспорт аналитики** — конец вехи 6, по отдельному подтверждению.
- **Наблюдаемость** (Prometheus/Grafana, RED-метрики, `/metrics`, алертинг) — убрано полностью 2026-08-25 (было E6.6). Причина: без стейджинга/прода и реального трафика метрики некому скрейпить и не на что алертить — мёртвая инфраструктура на MVP. NFR (p95≤300мс) на вехах 0-5 проверяется вручную (`hey`-прогон, задача 31 и её аналоги на других вехах), не постоянным мониторингом.
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

### Ревизия 2026-08-25 (go-reviewer аудит, вердикт REQUEST_CHANGES — 10 критичных находок, все закрыты)

| # | Находка | Решение |
|---|---|---|
| 10 | `catalog.institutions` без контактов/описания/адреса/обложки (FR-06 не покрыт) | Добавлены `description`/`address`/`phone`/`email`/`website`/`socials`/`cover_photo_s3_key`/`age_range`/`tag` |
| 11 | 8 метрик рейтинга негде взять при чтении карточки (`catalog` не вправе читать схему `reviews`) | `RatingSync.Apply` расширен до массива метрик + новая таблица `catalog.institution_metrics` |
| 12 | `auth.children` без UNIQUE — обходит единственную защиту от накрутки (FR-15/18) | `UNIQUE(user_id, institution_id)` |
| 13 | Ложное утверждение в плане «FK гарантирует approved-статус» (E2.6) | Формулировка исправлена: FK — ссылочная целостность, guard `moderation_status='approved'` — явный код в usecase |
| 14 | Родитель может привязать ребёнка к нескольким учреждениям с разными статусами; учреждение должно подтверждать привязку | Явно задокументировано (мульти-институция уже поддержана `UNIQUE` на пару, не на пользователя); добавлены `confirmation_status`/`confirmed_by`/`confirmed_at` на `auth.children`, подтверждение по личности родителя (не ребёнка), новый эндпоинт в кабинете учреждения, eligibility отзыва (E4.2) требует `confirmed` |
| 15 | `communications.messages` без отправителя и conversation — чат нереализуем | Новая `communications.conversations` + `messages.sender_type`/`sender_id` + раздельные `*_last_read_at` |
| 16 | FR-40 реализован частично: нет подтверждения email/согласия на ПД/сброса пароля; SMS дорогой для MVP | Флип приоритета: email/Google — основные каналы, телефон — опционален вне MVP; добавлены `auth.oauth_identities`, `auth.verification_codes`, `consent_at`/`consent_version`; `docs/EduHub_Functional_Requirements.md` синхронизирован (FR-40), SRS `.docx` — отложено |
| 17 | Нет механизма удаления аккаунта (закон РТ №1537) | Soft-delete + анонимизация вместо физического `DELETE` (`auth.users.status='deleted'`, затирание PII, отдельные фоновые задачи анонимизации по схемам-владельцам) |
| 18 | `Idempotency-Key` на `/auth/register` физически невозможен (`user_id NOT NULL`); гонка параллельных повторов не закрыта | `platform.idempotency_keys.user_id` → nullable; добавлен `status ∈ {in_progress,completed}`, INSERT как атомарная блокировка |
| 19 | Регистрация учреждения "идемпотентна" без соответствующего UNIQUE в схеме | Частичный `UNIQUE(lower(name->>'ru'), region, district) WHERE moderation_status<>'rejected'` на `catalog.institutions` |
| 20 | «Заявка на визит» есть на фронте, нет в схеме/плане | Новая `communications.visit_requests` (веха 5), идемпотентно + rate-limit |
| 21 | FR-25 (буст тарифа) конфликтует с keyset-пагинацией (курсор вехи 1 не учитывает буст в сортировке) | Продуктовое решение: буст позиции убран полностью, тариф не влияет на поиск вообще — конфликт снят исчезновением причины, не техническим патчем. Тариф теперь продаёт utility (лимиты/аналитика/охват), полный список — `docs/EduHub_Pricing_Tiers.md`. FR-25 в FR.md синхронизирован. Новое в схему (веха 3): `catalog.institutions.plan_expires_at`, `POST /api/v1/admin/institutions/{id}/plan` с audit-записью, per-plan лимиты в конфиге |
| 22 | FR-35 (спор по отзыву) недоукомплектован — критерий «на момент публикации» проверялся бы по текущему состоянию `auth.children`, не по снимку | `reviews.reviews.verified_at_publish` (снимок), `status` расширен до `disputed→resolved_kept\|resolved_removed`, добавлены `disputed_by`/`dispute_reason` |
| 23 | Отзывы об учреждении как о работодателе (Glassdoor-стиль) — новый функционал по явному запросу, ранее отложенный в 7-й волне фронта (ломал верификацию/метрики/монетизацию) | Все 3 причины отказа сняты дизайном: верификация трудоустройства модератором (не учреждением — иначе цензура своих же негативных отзывов), полностью отдельные метрики/агрегат (`reviews.employer_reviews`+`employer_review_metrics`, схема `auth.employment_claims`), видимость жёстко ограничена соискателями — учреждение никогда не видит на родительской витрине. FR-41 добавлен в FR.md |
| 24 | FR-34: нет хранилища материалов верификации владельца — только булев результат, без документа/обоснования решения | `catalog.institution_owner_verifications` (документ/тип документа/заметки модератора/история попыток). Уточнено фактом закона РТ: госшколы/сады лицензированию не подлежат вообще, лицензия нужна только частным организациям — `document_type` включает `state_status_confirmation` и `manual_exception` (обязательное текстовое обоснование модератора для центров без формальных документов), не только `license` |
| 25 | FR-07: presigned PUT в S3 обходит серверное автосжатие фото, нет обработки «осиротевших» загрузок (запрошен URL, файл не загружен) | Заменено на multipart-загрузку через бэкенд (E3.6, E5.6) — сервер видит байты (может сжать), запись в БД создаётся только после успешной загрузки, «осиротевших» ссылок не бывает по конструкции. Выбран вариант A из двух рассмотренных (альтернатива — presigned + confirm-эндпоинт + TTL-чистка — сложнее, не оправдано на объёме MVP) |
| 26 | FR-22/23: в событиях аналитики нет источника перехода — дашборд не может показать «источники трафика», хотя FR требует | `analytics.profile_events.referrer_class` (`search`/`catalog`/`map`/`direct`/`external`) — классификация `Referer` на бэкенде, сырой URL не хранится (тот же принцип, что `user_agent_class`, — полный referrer может нести PII в query источника) |
| 27 | FR-20 сведён к in-app без пометки, что push в PWA откладывать не нужно (в отличие от SMS/email — бесплатный встроенный браузерный стандарт, не внешний платный провайдер) | Push — реальный канал MVP (веха 5): `communications.push_subscriptions` (Web Push API, VAPID), `auth.users.notification_channels`. SMS/email остаются заглушками — там задержка объяснима |
| 28 | FR-01: `curriculum`/`program_level`/`discount` — индексы уже в задаче 18, но не попали в контракт эндпоинта вехи 1 | Добавлены в query-контракт (задачи 17/20/21/26) — SQL `&&` (пересечение), не `@>` (содержит все) — FR-01 буквально требует пересечение множеств |
| 29 | Rate-limiting спроектирован один раз (вехa 0, `/auth/login`), не зафиксирован явно для остальных эндпоинтов, принимающих пользовательский ввод | Явная таблица покрытия в разделе «Сквозные конвенции» — код подтверждения, чат, загрузка медиа, создание отзыва (родительского и о работодателе), заявка на визит, обращение к соискателю. Механизм не меняется — переиспользуется уже спроектированный |
| 30 | Два несогласованных механизма конкурентности (`version INT` vs ETag) упомянуты как взаимозаменяемые в E3.4 | `version`-поле убрано полностью. Единственный механизм — ETag/`updated_at` (уже есть в схеме, уже используется для 304 на чтении, задача 28) — `If-Match` на запись сверяется с тем же значением. Один механизм на оба направления, не два несинхронизированных |
| 31 | Двуязычный `JSONB{ru,tg}` там, где текст пишет живой человек в моменте (`reviews.reviews.text`/`reply`, `employer_responses.message`) — вынуждает дублировать или оставлять половину `null` | `TEXT`, любой язык независимо от языка интерфейса платформы — без `lang`-поля тоже (не добавлено, т.к. никакая функциональность на него не завязана). Тот же принцип, что уже был у `communications.messages.body` и `reviews.employer_reviews.text` |
| 32 | `auth.users` без имени и предпочитаемой локали — фронт показывает автора отзыва, уведомления некому языку локализовать | `display_name TEXT NULL` (не строгое ФИО, вводит сам пользователь, `NULL`→«Родитель» по умолчанию — минимизация PII) + `locale` (`ru`/`tg`, для локализации push/email) |
| 33 | Достижения соискателя (`web/lib/data.ts:2002`) негде хранить — `catalog.achievements` не покрывает `applicant`, а добавить туда нарушило бы владение схемами | `communications.applicant_achievements` — продублированная структура (не переиспользование чужой схемы), тот же паттерн что `employer_review_metrics`/`review_metrics` |
| 34 | Наблюдаемость (E6.6) приезжает только в вехе 6, вехи 0-5 без метрик | Пересмотрено и убрано полностью, не перенесено раньше — без стейджинга/прода метрики некому скрейпить, мёртвая инфраструктура. NFR на вехах 0-5 — ручной `hey`-прогон (задача 31 и аналоги), не мониторинг. Перенесено в Out of scope |
| 35 | Нет деплой-артефакта: `Dockerfile` не создаётся ни одной задачей, CI не собирает образ, MinIO нет в compose (вехи 3/5 пишут в S3 после перехода на multipart) | `docker-compose.yml` (задача 2) + MinIO; `Dockerfile` (multi-stage, distroless) и `docker build` в CI — присоединено к задаче 16, не отдельным номером (избежали коллизии нумерации с вехой 1) |
| 36 | Чат без правил доступа — любой пользователь может открыть диалог с кем угодно без ограничений (спам-вектор); плюс явный запрос: родители общаются друг с другом, не только с учреждением | `communications.conversations` обобщена — полиморфные `participant_a/b_type ∈ {user,institution}`, канонизация порядка в usecase (не в БД). Отдельный rate-limit на создание диалога (за период, не на сообщение). Механизм обнаружения собеседника-родителя — открытый вопрос, отдельная задача до вехи 5 |

**Закрыто в ревизии 2026-08-25** (решения #10-25 в таблице выше): контакты/описание институции, per-metric RatingSync, `auth.children` UNIQUE, FK-миф в E2.6, мультиучреждение+подтверждение учреждением, чат conversations/sender_type, FR-40 email/Google-приоритет+удаление аккаунта, idempotency race-fix, natural-key UNIQUE на регистрации, `visit_requests`, роль «Администратор платформы» (МОН убран из продукта целиком), FR-25 буст убран → utility-тарифы (`docs/EduHub_Pricing_Tiers.md`), FR-35 снимок верификации на момент публикации, отзывы о работодателе (FR-41), хранилище материалов FR-34, presigned→multipart-загрузка с серверным сжатием.

**Остаётся отложенным на отдельный проход** (не критичные находки go-reviewer): ограниченная перф-стратегия (только веха 1), версионирование API без политики, пожизненная уникальность `employer_responses` vs лимит за период, `moderation.queue_items` без UNIQUE на активный target, идемпотентность консьюмера «по event_id» без dedup-стора, партиционирование `analytics.profile_events` без джоба управления + `DROP PARTITION` (неверный синтаксис, нужен `DETACH`+`DROP TABLE`), критерий готовности вехи 3 vs назначение `audit_log`.

**Релевантные файлы фронта для сверки контрактов:** `web/lib/data.ts` (`Institution:191`, `Review:1184`, `Vacancy:1611`, `Applicant:1993`, `Application:2128`, `EmployerResponse:2112`, метрики `822-829`), `web/lib/app-state.tsx` (`ChildLink:21`, `RegisterInstitutionInput:31`), `web/app/(site)/search/page.tsx:56-116` (фильтры каталога), `web/lib/geo.ts` (клиентское определение координат).
