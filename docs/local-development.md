# Локальная разработка EduHub

Два способа запуска: **всё в Docker** (проще, ниже) и **бэк/фронт на хосте + только инфра в Docker** (быстрее для активной разработки Go/Next напрямую без контейнера).

## Способ 1 — всё в Docker (hot-reload)

Поднимает весь стек одной командой: PostgreSQL+PostGIS, Redis, MinIO, миграции, backend (Go, hot-reload через `air`), frontend (Next.js, `next dev`).

### Требования
- Docker Desktop с Docker Compose v2.20+ (директива `include` в `docker-compose.yml`)
- Свободные порты на хосте: `3000` (web), `8080` (api), `5433` (db), `6380` (redis), `9000`/`9001` (minio)

### Запуск
```bash
# из корня репозитория
docker compose up --build
# или в фоне:
docker compose up --build -d
```

Первый запуск — минуты (тянет образы `golang`, `node`, `postgis`, `redis`, `minio`). Дальше — секунды (кэш слоёв + именованные volumes для Go module/build cache).

### Что где
| Сервис | URL / адрес | Назначение |
|---|---|---|
| `web` | http://localhost:3000 | Next.js фронт (mock-данные, к API пока не подключён) |
| `api` | http://localhost:8080 | Go backend, `/healthz`, `/readyz` |
| `db` | localhost:5433 | PostgreSQL+PostGIS (внутри сети контейнеров — `db:5432`) |
| `redis` | localhost:6380 | Redis (внутри сети — `redis:6379`) |
| `minio` | http://localhost:9000 (API), :9001 (консоль) | S3-совместимое хранилище, логин/пароль `eduhub`/`eduhub12345` |

### Hot-reload
- **Backend**: правишь любой `.go` файл в `backend/` — `air` внутри контейнера `api` сам пересобирает и перезапускает бинарь (секунды). Смотреть процесс: `docker compose logs -f api`.
- **Frontend**: правишь что угодно в `web/` — Next.js (Turbopack) сам обновляет страницу в браузере (HMR), без перезапуска контейнера.

### Миграции
Применяются автоматически сервисом `migrate` при `docker compose up` (одноразовый контейнер, завершается после успеха). Проверить: `docker compose ps migrate` → `Exited (0)`.

Накатить вручную (например, после `docker compose up` уже поднятого стека, если добавил новую миграцию):
```bash
docker compose run --rm migrate
```
Откатить последнюю / посмотреть статус — тот же образ, `--entrypoint` переопределяет только бинарь, аргументы после имени сервиса передаются как есть (DSN в кавычках — `?` иначе раскроется шеллом как glob):
```bash
docker compose run --rm --entrypoint go migrate tool goose -dir migrations postgres "postgres://eduhub:eduhub@db:5432/eduhub?sslmode=disable" down
docker compose run --rm --entrypoint go migrate tool goose -dir migrations postgres "postgres://eduhub:eduhub@db:5432/eduhub?sslmode=disable" status
```

### Демо-данные
После миграций сервис `seed` автоматически прогоняет `cmd/devseed` (9 демо-институций из `web/lib/data.ts`) — тоже одноразовый контейнер, `api` ждёт его завершения. Проверить: `docker compose ps seed` → `Exited (0)`. Идемпотентен — безопасно перезапускать (`docker compose up` повторно, или вручную `docker compose run --rm seed`), дублей не будет.

### Остановка
```bash
docker compose down          # остановить + удалить контейнеры, volumes (БД, node_modules-volume и т.д.) остаются
docker compose down -v       # + снести volumes (потеряешь данные БД — только если реально нужно с нуля)
```

### Частые проблемы
- **`port is already allocated`** — что-то уже слушает порт (другой проект, старый `docker compose` стек, `npm run dev`/`make run` вне докера). Найти: `lsof -nP -iTCP:<порт> -sTCP:LISTEN` или `docker ps` (для чужих контейнеров). Освободить конфликтующий процесс/контейнер и повторить `docker compose up`.
- **`api`/`db` не видят друг друга по имени, DNS-ошибка `lookup db: no such host`** — обычно артефакт недособранного контейнера после прерванного `up`. Чинится пересозданием: `docker compose rm -f -s db && docker compose up -d`.
- **Правки в `.go` не подхватываются** — проверь, что volume реально смонтирован: `docker compose logs api | grep watching` должен перечислять пакеты. Если пусто — пересобрать: `docker compose up --build`.

---

## Способ 2 — бэк/фронт на хосте, инфра в Docker

Для тех, кто не хочет гонять всё через контейнеры (быстрее IDE-интеграция, дебаггер и т.п.).

### Backend
```bash
cd backend
cp .env.example .env               # один раз
docker compose up -d db redis minio # только инфра (без include, локальный backend/docker-compose.yml)
make migrate-up
make run                            # go run ./cmd/api → :8080
```
`.env` в этом случае указывает на `localhost:5433` (хостовый маппинг), не `db:5432` — то и другое уже настроено в `.env.example`.

### Frontend
```bash
cd web
npm install
npm run dev                         # → localhost:3000
```

### Проверка
```bash
make -C backend test-race
make -C backend lint
make -C backend vuln
```

---

## Смешивать способы 1 и 2 нельзя одновременно
Оба претендуют на одни и те же порты (`3000`, `8080`, `5433`, `6380`, `9000`/`9001`). Перед переключением — останови другой способ (`docker compose down` из корня, либо `Ctrl+C` у `make run`/`npm run dev` + `docker compose down` в `backend/`).

## Связка frontend↔backend
Сейчас фронт живёт на mock-данных (`web/lib/data.ts`) — реальные fetch-вызовы к API появятся в веху 1 (Catalog read-path). Инфраструктура для связки уже готова:
- `NEXT_PUBLIC_API_URL` прокинут в `web` (значение `http://localhost:8080` — адрес с точки зрения браузера)
- `CORS_ALLOWED_ORIGINS=http://localhost:3000` разрешает браузеру ходить с фронта на бэк (`httpx.CORS` middleware в `main.go`)
