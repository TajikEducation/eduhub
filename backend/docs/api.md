# EduHub API

Актуально на веху 1 (Catalog read-path) + платформенный скелет вехи 0. Обновлять вместе с
кодом при добавлении новых эндпоинтов (веха 2+). Формальный OpenAPI-спек не ведём — по
явному решению, пока эндпоинтов немного проще держать это как обычный markdown.

Базовый URL локально: `http://localhost:8080` (порт из `docker-compose.yml`/`.env.example`).
Postman-коллекция с рабочими примерами против реальных сидовых данных — `backend/postman/`.

## Общие конвенции

- Все ответы — `Content-Type: application/json; charset=utf-8`.
- Двуязычные поля (`{ru,tg}`) — везде, где контент готовит платформа/учреждение заранее
  (название, описание и т.д.), сериализуются как `{"ru": "...", "tg": "..."}`.
- Пустой список — всегда `[]`, никогда `null`.
- Каждый ответ несёт заголовок `X-Request-Id`; тело ошибки дублирует то же значение в
  `request_id`. Формат — сам request_id GET-запроса ничего не подтверждает клиенту, это
  диагностический идентификатор для логов, не токен идемпотентности.
- Пагинация каталога — непрозрачный курсор (`next_cursor`), не `offset`. Клиент не должен
  пытаться интерпретировать его содержимое, только передавать как есть в `cursor` следующего
  запроса. Внутри — `base64(JSON)` с позицией последней строки текущей страницы
  (`{"sort":"...","last_value":...,"last_id":"..."}`), следующая страница строится через
  `WHERE (поле_сортировки, id) > (last_value, last_id)` (keyset), не `OFFSET N` — не
  деградирует на росте каталога и не даёт дублей/пропусков при параллельных вставках между
  запросами страниц. `sort` внутри курсора обязан совпадать с `sort` запроса, которым он
  передаётся — курсор, выданный для `sort=score`, отклонится (`400`) при `sort=price_asc`.
- Тело запроса нигде в вехах 0-1 не требуется — оба эндпоинта каталога только `GET`, все
  параметры через query-строку/path.

### Формат ошибки

Единый для всех эндпоинтов и всех кодов ошибок:

```json
{
  "error": {
    "code": "invalid_input",
    "message": "человекочитаемое сообщение",
    "fields": { "sort": "недопустимое значение сортировки" }
  },
  "request_id": "9f1a17c1-066d-45da-940f-14acc41d7264"
}
```

`fields` присутствует только у `invalid_input` (per-field детали валидации) — у остальных
кодов его нет вообще (не `null`, ключ отсутствует, `omitempty`). Для `internal` (500)
`message` — всегда фиксированный текст `"internal server error"`, реальная причина никогда
не уходит наружу (логируется на сервере под тем же `request_id`).

| HTTP-статус | `code` | Когда |
|---|---|---|
| 400 | `invalid_input` | Синтаксис/формат query-параметра или тела не прошёл валидацию |
| 401 | `unauthorized` | Маппинг существует в `httpx`, но ни один текущий эндпоинт его не возвращает — нет ни одного auth-защищённого роута (появится с вехой 2) |
| 403 | `forbidden` | Аналогично — маппинг есть, эндпоинтов, которые его отдают, пока нет |
| 404 | `not_found` | Ресурс не найден (включая `pending`/`rejected` институции — публичный каталог не раскрывает сам факт их существования) |
| 409 | `conflict` | Маппинг есть, мутирующих эндпоинтов, которые могли бы его вернуть, пока нет (веха 3) |
| 429 | `rate_limited` | In-memory token-bucket лимитер реализован (задача 15) и покрыт тестами, но **пока не подключён** ни к одному реальному роуту в `cmd/api/router.go` — сработает только когда будет явно добавлен в цепочку middleware конкретного эндпоинта |
| 500 | `internal` | Непредвиденная ошибка сервера |

---

## Health

### `GET /healthz`

Liveness — процесс жив, без обращения к БД/Redis/чему-либо ещё.

**Запрос**
```
GET /healthz
```

Входных параметров нет (ни query, ни path, ни тела).

**Ответы**

Единственный вариант — `200`, пока процесс вообще отвечает на запросы (нет ветки отказа):

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Request-Id: 3938fb84-a645-4dba-a04b-2a7199cf0e3b

{"status": "ok"}
```

| Поле | Тип | Nullable | Описание |
|---|---|---|---|
| `status` | string | нет | Всегда `"ok"` |

### `GET /readyz`

Readiness — проверяет каждую зарегистрированную зависимость (сейчас: Postgres) в пределах
внутреннего таймаута 2с. Redis **намеренно не входит** в эту проверку — кэш деградирует
(прямое чтение из БД), недоступность Redis не делает сервис неготовым.

**Запрос**
```
GET /readyz
```

Входных параметров нет.

**Ответы**

Все зависимости отвечают:

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{"status": "ok"}
```

Хотя бы одна упала или не успела за 2с — статус `503`, тело называет упавшие по имени:

```http
HTTP/1.1 503 Service Unavailable
Content-Type: application/json; charset=utf-8

{"status": "unavailable", "failed": ["db"]}
```

| Поле | Тип | Nullable | Описание |
|---|---|---|---|
| `status` | string | нет | `"ok"` или `"unavailable"` |
| `failed` | string[] | да (только при `503`) | Имена упавших зависимостей (сейчас единственно возможное значение — `"db"`) |

---

## Каталог учреждений

### `GET /api/v1/institutions`

Список учреждений с фильтрами, сортировкой и keyset-пагинацией. Публичный каталог всегда
показывает только `moderation_status=approved` — форсируется сервисом, никакой query-параметр
на это не влияет.

#### Query-параметры

| Параметр | Тип | Описание |
|---|---|---|
| `q` | string | Подстрочный поиск по названию (`ru`/`tg` одновременно, `pg_trgm`) |
| `type` | string, через запятую | Фильтр по типу(ам) учреждения (`kindergarten`\|`preschool`\|`school`\|`center`\|`university`) |
| `region` | string | `dushanbe`\|`sughd`\|`khatlon`\|`gbao`\|`rrp` |
| `area` | string | Район/микрорайон (точное совпадение `district`) |
| `min_price` / `max_price` | int | Диапазон цены |
| `min_rating` | float | Институции с `rating_avg IS NULL` исключаются (нет данных ≠ ноль) |
| `transport` / `food` / `verified` / `discount` | bool (`true`/`false`) | Отсутствие параметра ≠ `false` — это разные состояния фильтра (`*bool`) |
| `curriculum` | string, через запятую | Пересечение (институция матчит хотя бы по одному значению), не точное совпадение |
| `program_level` | string, через запятую | Та же семантика пересечения |
| `sort` | string | `""` (дефолт, новые сначала по `created_at`) \| `price_asc` \| `score`. Другие значения (например `price_desc`, `reviews`) пока не реализованы. |
| `lat` / `lng` | float | Координаты для гео-поиска — обязательны вместе, диапазоны `[-90,90]`/`[-180,180]` |
| `radius_km` | float | Требует заданных `lat`/`lng` |
| `limit` | int | По умолчанию 20, максимум 50 |
| `cursor` | string | Значение из `next_cursor` предыдущего ответа |

Все параметры опциональны — `GET /api/v1/institutions` без единого query-параметра валиден
и возвращает первую страницу дефолтного порядка.

#### Запрос

```
GET /api/v1/institutions?region=dushanbe&curriculum=bilingual,state&min_price=100&max_price=1000&sort=price_asc&limit=20
```

Пример с гео (координаты обязаны идти парой):
```
GET /api/v1/institutions?lat=38.5598&lng=68.7870&radius_km=15&sort=score
```

Пример постраничного продолжения:
```
GET /api/v1/institutions?limit=20&sort=score&cursor=eyJzb3J0Ijoic2NvcmUiLCJsYXN0X3ZhbHVlIjo0LjcsImxhc3RfaWQiOiIuLi4ifQ==
```

#### Ответы

**`200` — есть результаты**, заголовки несут `Cache-Control` (публично кэшируемо 60с):

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: public, max-age=60
X-Request-Id: dbf482a4-3cf3-455c-b661-f737185339bf

{
  "items": [
    {
      "id": "c96d25ee-40ec-4fd6-ba0f-c0a7409236b3",
      "name": {"ru": "Центр «Ояндасоз»", "tg": "Маркази «Ояндасоз»"},
      "types": ["center"],
      "region": "khatlon",
      "city": {"ru": "Бохтар", "tg": "Бохтар"},
      "district": "Бохтар",
      "price": 300,
      "rating_avg": 4.2,
      "review_count": 41,
      "verified": false,
      "discount_available": false,
      "cover_photo_s3_key": "https://...",
      "tag": {"ru": "Новинка", "tg": "Наваҷӯ"},
      "distance_m": 577.17
    }
  ],
  "next_cursor": "eyJzb3J0Ijoic2NvcmUiLCJsYXN0X3ZhbHVlIjo0LjcsImxhc3RfaWQiOiIuLi4ifQ==",
  "total_hint": null
}
```

Карточка списка — сокращённая (без сателлитных коллекций и профильных полей типа `phone`).
`distance_m` присутствует только когда `lat`/`lng` заданы в запросе (`omitempty` на `nil`, не
на `0` — отсутствие гео и «расстояние 0» различимы). `total_hint` всегда `null` — keyset-
пагинация не делает `COUNT`-запрос. Отсутствующие optional-поля (например `city`/`tag`, если
`NULL` в БД) не сериализуются вовсе, как и в карточке `{id}` ниже.

**Поля ответа**

| Поле | Тип | Nullable (`omitempty`) | Описание |
|---|---|---|---|
| `items` | array | нет (`[]`, не `null`, если пусто) | Страница институций |
| `items[].id` | UUID | нет | — |
| `items[].name` | `{ru,tg}` | нет | — |
| `items[].types` | string[] | нет (может быть `[]`) | Один или несколько типов учреждения |
| `items[].region` | string | нет | `dushanbe`\|`sughd`\|`khatlon`\|`gbao`\|`rrp` |
| `items[].city` | `{ru,tg}` | да | — |
| `items[].district` | string | да | — |
| `items[].price` | int | да | Стоимость, если указана учреждением |
| `items[].rating_avg` | float | да | `NULL` — «нет данных», не 0 |
| `items[].review_count` | int | нет | `0`, если отзывов нет (не путать с `NULL` у `rating_avg`) |
| `items[].verified` | bool | нет | — |
| `items[].discount_available` | bool | нет | — |
| `items[].cover_photo_s3_key` | string | да | — |
| `items[].tag` | `{ru,tg}` | да | Например «Топ выбор» |
| `items[].distance_m` | float | да | Только если `lat`/`lng` заданы в запросе |
| `next_cursor` | string | да (`null`, если страница последняя) | См. «Общие конвенции» выше |
| `total_hint` | int | всегда `null` этой вехой | Зарезервировано, не заполняется (нет `COUNT`-запроса) |

**`200` — ничего не найдено** (не `404` — пустой список валидных результатов, не ошибка):

```json
{"items": [], "next_cursor": null, "total_hint": null}
```

**`400` — невалидный `sort`:**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{
  "error": {"code": "invalid_input", "message": "sort невалиден", "fields": {"sort": "недопустимое значение сортировки"}},
  "request_id": "873834d1-d09d-44c7-ba59-edaf5dde5db8"
}
```

**`400` — `lat` без `lng`** (тот же код, другое поле в `fields`):

```json
{
  "error": {"code": "invalid_input", "message": "lng отсутствует", "fields": {"lng": "обязателен вместе с lat"}},
  "request_id": "81a809d3-dd09-4ef4-a699-c91d6bdbc06a"
}
```

Тот же класс `400 invalid_input` (с соответствующим полем в `fields`) возвращается и на:
`min_price`/`max_price`/`limit`/`min_rating`/`radius_km`, если значение не парсится как
число; `transport`/`food`/`verified`/`discount`, если значение не парсится как bool;
`min_price > max_price`; `radius_km` без `lat`/`lng`; `lat` вне `[-90,90]` или `lng` вне
`[-180,180]`; битый (не-base64 или не тот `sort`, для которого выдан) `cursor`.

### `GET /api/v1/institutions/{id}`

Полная карточка учреждения (профильная страница, FR-06) — все скалярные поля + сателлитные
коллекции одним SQL-запросом.

#### Входные параметры

| Параметр | Расположение | Тип | Обязателен | Описание |
|---|---|---|---|---|
| `id` | path | UUID | да | Невалидный формат → `400`, не найден/не approved → `404` |
| `If-None-Match` | header | string | нет | Значение `ETag` предыдущего ответа — совпадение даёт `304` |

#### Запрос

```
GET /api/v1/institutions/c96d25ee-40ec-4fd6-ba0f-c0a7409236b3
```

Условный повторный запрос (см. кэширование ниже):
```
GET /api/v1/institutions/c96d25ee-40ec-4fd6-ba0f-c0a7409236b3
If-None-Match: "c96d25ee-40ec-4fd6-ba0f-c0a7409236b3.1788074490060302000"
```

#### Ответы

**`200`** — все optional-поля (`*T` в Go) `omitempty`: при `NULL` в БД поле **отсутствует в
JSON целиком**, не сериализуется как `null`. Реальный пример (сокращён — `staff`/`gallery`
урезаны до одной записи, у этой конкретной институции нет `achievements`/`alumni`/
`transport_routes`/`meal_plans`/`socials`/`license_no`/`languages`/`program_level`/
`curriculum`/`discount_type`/`discount_details`/`location_landmarks` — они просто не
появляются в ответе):

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Etag: "c96d25ee-40ec-4fd6-ba0f-c0a7409236b3.1788074490060302000"
X-Request-Id: 17b336af-ccdd-4a27-b6d6-942cbb25073a

{
  "id": "c96d25ee-40ec-4fd6-ba0f-c0a7409236b3",
  "name": {"ru": "Центр «Ояндасоз»", "tg": "Маркази «Ояндасоз»"},
  "types": ["center"],
  "region": "khatlon",
  "city": {"ru": "Бохтар", "tg": "Бохтар"},
  "district": "Бохтар",
  "description": {"ru": "...", "tg": "..."},
  "address": {"ru": "ул. Исмоили Сомони 12, Бохтар, Хатлонская область", "tg": "..."},
  "lat": 37.8322,
  "lng": 68.7803,
  "phone": "+992 91 500-40-10",
  "email": "info@oyandasoz.tj",
  "website": "oyandasoz.tj",
  "cover_photo_s3_key": "https://...",
  "age_range": "любой",
  "tag": {"ru": "Новинка", "tg": "Наваҷӯ"},
  "price": 300,
  "discount_available": false,
  "verified": false,
  "founded": 2018,
  "students_count": 95,
  "rating_avg": 4.2,
  "review_count": 41,
  "created_at": "2026-08-30T07:21:30.060302Z",
  "updated_at": "2026-08-30T07:21:30.060302Z",
  "staff": [
    {
      "id": "05bc51fb-ce66-4450-b489-f54218585398",
      "name": {"ru": "Сафармад Кутбиддинов", "tg": "..."},
      "role_type": "teacher",
      "role_label": {"ru": "Преподаватель математики", "tg": "..."},
      "subject": {"ru": "Математика", "tg": "Математика"},
      "photo_url": "https://...",
      "exp": "8 лет",
      "bio": {"ru": "...", "tg": "..."},
      "email": "s.kutbiddinov@oyandasoz.tj"
    }
  ],
  "achievements": [],
  "gallery": [
    { "id": "a13d9be0-0964-4244-9082-2cf9a6b479f1", "s3_key": "https://...", "label": {"ru": "Занятия", "tg": "..."}, "sort_order": 0 }
  ],
  "alumni": [],
  "transport_routes": [],
  "meal_plans": []
}
```

Сознательно **не включено ни при каких условиях** (не просто `omitempty` — этих полей нет в
DTO вообще): `moderation_status` (внутреннее — `Get` уже гарантирует только `approved`),
`plan`/`plan_expires_at` (биллинговые, не публичные), `news` (репозиторий пока не собирает
новости в карточку — задача 24), разбивка по 8 метрикам рейтинга (`catalog.institution_metrics`
пока пустая, заполнится вехой 4).

**Поля ответа — скаляры**

| Поле | Тип | Nullable (`omitempty`) | Описание |
|---|---|---|---|
| `id` | UUID | нет | — |
| `name` | `{ru,tg}` | нет | — |
| `types` | string[] | нет | — |
| `region` | string | нет | — |
| `city` | `{ru,tg}` | да | — |
| `district` | string | да | — |
| `description` | `{ru,tg}` | да | — |
| `address` | `{ru,tg}` | да | Полный адрес, отдельно от `city`/`district` |
| `lat` / `lng` | float | нет | Координаты `GEOGRAPHY(Point)`, всегда присутствуют |
| `location_landmarks` | string | да | Ориентиры («рядом с рынком») |
| `phone` / `email` / `website` | string | да | — |
| `socials` | object | да | `{instagram,telegram,facebook}`, каждое поле внутри тоже `omitempty` |
| `cover_photo_s3_key` | string | да | Обложка, ключ/URL в хранилище медиа |
| `age_range` | string | да | Свободный текст («7–17 лет») |
| `tag` | `{ru,tg}` | да | — |
| `license_no` | string | да | Не заполняется этой вехой — нет источника данных |
| `languages` | string[] | да | — |
| `program_level` | string[] | да | — |
| `curriculum` | string[] | да | `state`\|`bilingual`\|`international`\|`stem` |
| `price` | int | да | — |
| `discount_available` | bool | нет | — |
| `discount_type` | string[] | да | — |
| `discount_details` | string | да | — |
| `verified` | bool | нет | — |
| `founded` | int | да | Год основания |
| `students_count` | int | да | — |
| `rating_avg` | float | да | `NULL` — «нет данных» |
| `review_count` | int | нет | — |
| `created_at` / `updated_at` | ISO 8601 timestamp | нет | `updated_at` — источник `ETag` |

**Поля ответа — вложенные коллекции** (каждая — `[]`, никогда `null`, если пуста)

| Коллекция | Поле элемента | Тип | Nullable | Описание |
|---|---|---|---|---|
| `staff[]` | `id` | UUID | нет | — |
| | `name` | `{ru,tg}` | нет | — |
| | `role_type` | string | нет | `director`\|`teacher`\|`coach`\|`psychologist`\|`admin` |
| | `role_label` | `{ru,tg}` | нет | Человекочитаемая должность |
| | `subject` | `{ru,tg}` | да | Предмет (для `role_type=teacher`) |
| | `photo_url` | string | да | — |
| | `exp` | string | да | Стаж, свободный текст |
| | `bio` | `{ru,tg}` | да | — |
| | `email` / `phone` | string | да | — |
| `achievements[]` | `id` | UUID | нет | Полиморфно: институции или сотрудника |
| | `title` | `{ru,tg}` | нет | — |
| | `year` | int | нет | — |
| | `category` | string | нет | `gold`\|`silver`\|`bronze`\|`special` |
| | `description` | `{ru,tg}` | нет | — |
| | `links[]` | `{label,url}` | нет (`[]`) | — |
| `gallery[]` | `id` | UUID | нет | — |
| | `s3_key` | string | нет | — |
| | `label` | `{ru,tg}` | да | Подпись фото |
| | `sort_order` | int | нет | Порядок отображения |
| `alumni[]` | `id` | UUID | нет | — |
| | `name` | `{ru,tg}` | нет | — |
| | `photo_url` | string | да | — |
| | `grad_year` | int | нет | — |
| | `now_label` | `{ru,tg}` | да | «Студент МФТИ, физика» |
| `transport_routes[]` | `id` | UUID | нет | — |
| | `type` | string | нет | `own_bus`\|`minibus`\|`taxi`\|`parent_coop`\|`other` |
| | `label` | `{ru,tg}` | да | Свободное описание маршрута |
| | `areas[]` | `{ru,tg}` | нет (`[]`) | Районы охвата |
| | `cost` | int | да | — |
| | `cost_period` | string | нет | `month`\|`day`\|`trip` |
| | `sort_order` | int | нет | — |
| `meal_plans[]` | `id` | UUID | нет | — |
| | `meal_type` | string | нет | `hot`\|`breakfast`\|`buffet`\|`other` |
| | `label` | `{ru,tg}` | да | — |
| | `cost` | int | да | — |
| | `cost_period` | string | нет | `month`\|`day` |
| | `halal` | bool | да | `NULL` = не указано, отличается от «не халяль» (`false`) |
| | `sort_order` | int | нет | — |

**`304 Not Modified`** — `If-None-Match` совпал с текущим `ETag`, тело пустое:

```http
HTTP/1.1 304 Not Modified
```

**`400`** — `{id}` не парсится как UUID:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8

{
  "error": {"code": "invalid_input", "message": "некорректный id", "fields": {"id": "невалидный UUID"}},
  "request_id": "..."
}
```

**`404`** — валидный UUID, но записи нет (или она `pending`/`rejected` — публичный каталог
не должен подтверждать сам факт существования немодерированной институции):

```http
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8

{
  "error": {"code": "not_found", "message": "institution not found: id=00000000-0000-0000-0000-000000000000"},
  "request_id": "8b2071ec-1d51-4b01-99aa-68a51fbe22ad"
}
```

#### Кэширование (`ETag`/`If-None-Match`)

Ответ несёт `ETag: "<id>.<updated_at в unix-нано>"`. Повторный запрос с `If-None-Match`,
равным этому значению → `304`. Значение детерминировано от `id`+`updated_at` — любая мутация
карточки (веха 3) обязана поднимать `updated_at`, иначе ETag не отразит изменение.

---

## Пока не реализовано

Веха 2 (Auth/RBAC) добавит `/auth/*` и `X-Request-Id`-совместимую авторизацию (`Authorization: Bearer`).
Вехи 3-6 добавят мутирующие эндпоинты каталога, отзывы, чат/вакансии, аналитику — контракты
появятся в этом файле по мере реализации, не заранее.
