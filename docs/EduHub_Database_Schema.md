# EduHub — модель данных бэкенда (схема БД)

**Дата:** 2026-08-25 (ревизия после go-reviewer аудита — 10 критичных находок закрыты)
**Статус:** согласовано, справочный документ (детализация `docs/EduHub_Backend_Development_Plan.md`)
**СУБД:** PostgreSQL + PostGIS, одна база `eduhub`, схемы по владению (см. `.claude/rules/architecture.md`, `.claude/rules/db.md`)

## Как читать этот документ

Каждая таблица — в своей схеме, пишет в неё только сервис-владелец схемы (`.claude/rules/architecture.md`). Кросс-схемные FK запрещены, кроме двух явно согласованных исключений (см. ниже). Все ID — `UUID`, всё время — `TIMESTAMPTZ`, двуязычные поля — `JSONB {ru, tg}` (симметрично `Bi` из `web/lib/data.ts:23`), гео — `GEOGRAPHY(Point,4326)` с `GIST`-индексом. Опциональные поля отмечены `NULL`; на Go-стороне им соответствует `*T`, не нулевое значение (`.claude/rules/go.md`).

**Согласованные кросс-схемные исключения:**
- `auth.children.institution_id → catalog.institutions(id)` — обычный физический FK, `ON DELETE RESTRICT`. Институции физически не удаляются ни в одном сценарии MVP, только меняют `moderation_status` — риска блокировки нет.
- Всё остальное между схемами — только через `user_id`/владельца или проверяется в service/usecase-слое (порты вроде `auth.ChildLinkExists`), не через FK.

---

## Схема `platform`

Инфраструктурные таблицы, не принадлежат ни одному бизнес-домену.

### `platform.idempotency_keys`

Хранит ответы на мутирующие запросы с `Idempotency-Key`, чтобы повтор (сетевой ретрай, двойной клик) не выполнял операцию дважды. `user_id` nullable — нужна поддержка анонимных эндпоинтов (например `/auth/register`, где на момент запроса пользователя ещё не существует).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `key` | TEXT | нет | — | значение заголовка `Idempotency-Key`, часть композитного ключа |
| `user_id` | UUID | **да** | `NULL` | кто сделал запрос — `NULL` для анонимных эндпоинтов (без FK — не критично для этой служебной таблицы) |
| `endpoint` | TEXT | нет | — | `METHOD /path`, чтобы один ключ не спутал разные операции |
| `request_hash` | TEXT | нет | — | хеш тела запроса — тот же ключ с другим телом = ошибка, не тихая подмена |
| `status` | TEXT | нет | `'in_progress'` | `in_progress`\|`completed` — строка вставляется **до** выполнения операции (атомарный "замок" против гонки параллельных повторов); при конфликте вставки: `completed` → отдать сохранённый ответ, `in_progress` → `409 Conflict` |
| `response_status` | INT | да | `NULL` | HTTP-код сохранённого ответа — заполняется по завершении |
| `response_body` | JSONB | да | `NULL` | тело сохранённого ответа для повторной отдачи — заполняется по завершении |
| `created_at` | TIMESTAMPTZ | нет | `now()` | для TTL-очистки через 24ч |

**Связи:** нет FK — намеренно, служебная таблица переживает удаление сущностей.
**Индексы:** `UNIQUE(key, endpoint)` (без `user_id` в ключе — иначе анонимные запросы с `user_id IS NULL` не считались бы дублями друг друга, `NULL` не равен `NULL` в UNIQUE); `btree(created_at)` для фоновой чистки.
**Инвариант конкурентности:** INSERT строки со статусом `in_progress` — атомарная операция, служит распределённой блокировкой; при `unique_violation` смотрим текущий статус существующей строки вместо повторного выполнения операции.

### `platform.seed_refs`

Служебная таблица идемпотентности `cmd/devseed` — повторный запуск сидера не плодит дубли демо-данных.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `seed_ref` | TEXT | нет | — | стабильный ключ из исходных mock-данных (например `"web-inst-1"`) |
| `entity_type` | TEXT | нет | — | `institution`/`staff`/`vacancy`/... |
| `entity_id` | UUID | нет | — | реальный UUID созданной записи |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет FK (сознательно — сидер не должен зависеть от чужих схем).
**Индексы:** `UNIQUE(seed_ref, entity_type)`.

---

## Схема `auth`

Владелец: сервис Auth (веха 2). Аккаунты, сессии, минимальная привязка ребёнка для верификации отзывов.

### `auth.users`

Один аккаунт на человека (SRS FR-40 — без ветки «кто вы» при регистрации). **Приоритет каналов пересмотрен 2026-08-25**: email/пароль + вход через Google — основные, телефон — опциональное поле второй очереди (SMS-провайдер в РТ платный за каждую отправку, на MVP с холодным стартом не оправдан). Отменяет исходную формулировку FR-40 в SRS v2.1 («телефон — основной путь»); требуется синхронизация `docs/EduHub_Functional_Requirements.md` (сделано в этой ревизии) и SRS `.docx` (отложено, см. план).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `email` | TEXT | нет | — | основной канал, нормализованный (lowercase/trim) |
| `display_name` | TEXT | да | `NULL` | имя, вводимое самим пользователем (не обязательно паспортное ФИО — как никнейм); `NULL` отображается как «Родитель» по умолчанию у отзыва (добавлено 2026-08-25) |
| `locale` | TEXT | нет | `'ru'` | `ru`\|`tg` — для локализации уведомлений (push/email) конкретному пользователю (добавлено 2026-08-25) |
| `phone` | TEXT | да | `NULL` | нормализованный формат `+992XXXXXXXXX`, опциональное поле второй очереди — SMS-верификация вне MVP |
| `password_hash` | TEXT | да | `NULL` | argon2id, PHC-строка; `NULL` для аккаунтов, заведённых только через Google (`oauth_identities`) |
| `role` | TEXT | нет | `'user'` | `guest`\|`user`\|`institution`\|`moderator`\|`admin` (CHECK) |
| `status` | TEXT | нет | `'unverified'` | `unverified`\|`active`\|`banned`\|`deleted` (CHECK) — `deleted` для soft-delete (закон РТ №1537, см. ниже) |
| `email_verified_at` | TIMESTAMPTZ | да | `NULL` | заполняется после подтверждения кодом (E2.4) |
| `consent_at` | TIMESTAMPTZ | нет | — | момент согласия на обработку ПД — юридическое доказательство по закону РТ №1537 |
| `consent_version` | TEXT | нет | — | версия текста политики, на которую согласился пользователь — при обновлении политики старое согласие не считается действительным на новую версию |
| `failed_login_count` | INT | нет | `0` | anti-bruteforce (E2.7) |
| `locked_until` | TIMESTAMPTZ | да | `NULL` | блокировка входа после N неудач |
| `notification_channels` | JSONB | нет | `'{"push":true,"email":true}'` | простой набор переключателей канала уведомлений (добавлено 2026-08-25) — не история, поэтому не отдельная таблица |
| `deleted_at` | TIMESTAMPTZ | да | `NULL` | момент soft-delete — при заполнении `email`/`phone`/`password_hash` затираются анонимизирующим значением |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** родитель для `refresh_tokens`, `children`, `oauth_identities`, `verification_codes`, и (по `user_id`, без физического FK за пределами схемы) для владельцев/акторов во всех остальных схемах.
**Индексы:** `UNIQUE(email)`; частичный `UNIQUE(phone) WHERE phone IS NOT NULL`.

**Примечание к RBAC:** «родитель» и «соискатель» — НЕ роли, а факты о пользователе (наличие записей в `children`, значение `applicants.visibility`) — решение волны 7 фронта, уже так реализовано в `web/lib/app-state.tsx`. `role` в этой таблице — только уровень доступа (RBAC-модель B).

**Удаление аккаунта (закон РТ №1537, право на удаление):** не физический `DELETE`, а soft-delete + анонимизация — `status='deleted'`, `deleted_at`, `email`/`phone`/`password_hash` затираются (например `deleted-<uuid>@eduhub.local`), все `refresh_tokens` отзываются разом. Данные в других схемах, ссылающиеся по значению `user_id` (`reviews.reviews`, `communications.applicants` и т.д.), анонимизируются отдельными фоновыми задачами каждой схемы-владельца — не каскадно, т.к. кросс-схемных FK нет; `communications.applicants` (самое PII-плотное место — имя/телефон/email/CV) затирается полностью, `cv_s3_key` файл удаляется из S3 отдельной задачей. `reviews.reviews.text` и `moderation.audit_log` НЕ анонимизируются — общественная ценность отзыва и законный интерес журнала модерации сохраняются, личность автора — нет.

### `auth.oauth_identities`

Вход через Google без пароля — привязка внешней идентичности к аккаунту.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID FK→`auth.users(id)` ON DELETE CASCADE | нет | — | — |
| `provider` | TEXT | нет | — | `'google'` (CHECK, единственное значение на MVP) |
| `provider_user_id` | TEXT | нет | — | `sub` из Google-профиля |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.users`.
**Индексы:** `UNIQUE(provider, provider_user_id)`.

### `auth.verification_codes`

Короткоживущие коды подтверждения — email (основной канал), место под будущее расширение на телефон.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID FK→`auth.users(id)` ON DELETE CASCADE | нет | — | — |
| `channel` | TEXT | нет | — | `email`\|`phone` (CHECK) |
| `purpose` | TEXT | нет | — | `register`\|`password_reset` (CHECK) |
| `code_hash` | TEXT | нет | — | хранится только хеш кода, не сам код |
| `attempts_count` | INT | нет | `0` | лимит попыток ввода |
| `expires_at` | TIMESTAMPTZ | нет | — | короткий TTL (минуты) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.users`.
**Индексы:** `btree(user_id, channel, purpose)`; частота отправки — Redis-счётчик по email+IP, не в этой таблице.

### `auth.employment_claims`

Заявка сотрудника (текущего/бывшего) на подтверждение трудоустройства — основа для отзыва об учреждении как о работодателе (`reviews.employer_reviews`, добавлено 2026-08-25). **Принципиально не как `auth.children`**: верификацию делает **модератор платформы**, не само учреждение — если бы учреждение подтверждало, оно могло бы просто не подтверждать авторов негативных отзывов, что убивает цель честного отзыва о работодателе.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID FK→`auth.users(id)` ON DELETE CASCADE | нет | — | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE RESTRICT | нет | — | тот же паттерн исключения, что у `children` |
| `position` | TEXT | да | `NULL` | должность, информационно |
| `employment_status` | TEXT | нет | — | `current`\|`former` (CHECK) |
| `verification_status` | TEXT | нет | `'pending'` | `pending`\|`verified`\|`rejected` (CHECK) — решает модератор, не учреждение |
| `evidence_s3_key` | TEXT | да | `NULL` | опциональный подтверждающий документ (трудовой договор/справка) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.users`; N—1 `catalog.institutions`.
**Индексы:** `UNIQUE(user_id, institution_id)`; `btree(verification_status)` (очередь модератора).

### `auth.refresh_tokens`

Ротация сессий с reuse-detection (E2.3).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID FK→`auth.users(id)` ON DELETE CASCADE | нет | — | — |
| `token_hash` | TEXT | нет | — | хранится только хеш, не сам токен |
| `family_id` | UUID | нет | — | все токены одной цепочки ротации несут общий id — при обнаружении reuse отзывается вся семья |
| `expires_at` | TIMESTAMPTZ | нет | — | +30 дней от выдачи |
| `revoked_at` | TIMESTAMPTZ | да | `NULL` | заполняется при logout или reuse-detection |
| `replaced_by` | UUID FK→`auth.refresh_tokens(id)` | да | `NULL` | на какой токен заменён при ротации |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.users`; самоссылка `replaced_by`.
**Индексы:** `btree(token_hash)`; `btree(family_id)`; `btree(user_id, expires_at)`.

### `auth.children`

SRS §7, сущность `Child` — минимальная привязка родитель↔учреждение для верификации отзыва (FR-15). **Явно НЕ хранит** имя, фамилию, фото, дату рождения, контакты ребёнка — минимизация PII. Таблица — привязок, не «детей»: **один пользователь может иметь несколько строк одновременно, по одной на каждое учреждение** (например ребёнок одновременно ходит в школу и в доп.центр — обе привязки `status='current'` в одно и то же время). `UNIQUE(user_id, institution_id)` ограничивает только повтор одной и той же пары, не общее число привязок пользователя.

**Подтверждение учреждением (добавлено 2026-08-25):** привязка — самодекларация родителя до момента подтверждения. Право на отзыв (E4.2, FR-15) требует не просто существования строки, а `confirmation_status='confirmed'`. Так как ребёнок в системе безымянный, подтверждение идёт **по личности родителя** (email/имя, которые родитель сам указал при регистрации — это его собственные ПД, не ребёнка), не по данным о ребёнке — секретарь/директор учреждения узнаёт заявителя по своим внутренним спискам и подтверждает или отклоняет заявку в кабинете учреждения.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID FK→`auth.users(id)` ON DELETE CASCADE | нет | — | родитель |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE RESTRICT | нет | — | **согласованное кросс-схемное исключение** — см. шапку документа. Ссылочная целостность (строка существует), НЕ бизнес-инвариант «институция одобрена» — тот проверяется отдельно в usecase при создании привязки, FK его не гарантирует |
| `age_group` | TEXT | нет | — | `kindergarten`\|`preschool`\|`primary`\|`basic`\|`secondary`\|`university` (CHECK) |
| `status` | TEXT | нет | — | `current`\|`alumnus`\|`transferred` (CHECK) — `transferred` для верификации отзыва (FR-15/30) трактуется как `alumnus`: связь была реальной |
| `confirmation_status` | TEXT | нет | `'pending'` | `pending`\|`confirmed`\|`rejected` (CHECK) — до `confirmed` привязка видна родителю в кабинете, но не даёт права на отзыв |
| `confirmed_by` | UUID | да | `NULL` | представитель учреждения, подтвердивший заявку — по значению, без FK |
| `confirmed_at` | TIMESTAMPTZ | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.users`; N—1 `catalog.institutions` (физический FK, исключение из правила схем).
**Индексы:** `UNIQUE(user_id, institution_id)` — защита от самонакрутки: без этого ограничения пользователь мог бы создать множество привязок к одному учреждению и обойти анти-фрод логику отзывов (FR-15/FR-18); `btree(institution_id, confirmation_status)` (для очереди подтверждения в кабинете учреждения и для eligibility-проверки отзыва в вехе 4 через порт `auth.ChildLinkExists`, которая теперь проверяет `confirmation_status='confirmed'`, не только факт существования строки).

---

## Схема `catalog`

Владелец: сервис Catalog (вехи 1 и 3). Учреждения и всё, что относится к их публичному профилю.

### `catalog.institutions`

Ядро платформы — учреждение (SRS §7.2, с расширениями из реального фронта `web/lib/data.ts:191-228`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `name` | JSONB | нет | — | `{ru,tg}` |
| `types` | TEXT[] | нет | — | мульти-тип (FR-27) — `kindergarten`\|`preschool`\|`school`\|`center`\|`university`, для комплексов несколько значений |
| `region` | TEXT | нет | — | `dushanbe`\|`sughd`\|`khatlon`\|`gbao`\|`rrp` (CHECK) |
| `city` | JSONB | да | `NULL` | `{ru,tg}` |
| `district` | TEXT | да | `NULL` | район |
| `description` | JSONB | да | `NULL` | `{ru,tg}` — полное описание, FR-06 «всё нужное, не переходя на внешние сайты» |
| `address` | JSONB | да | `NULL` | `{ru,tg}` — текстовый адрес (в дополнение к `geo`, для отображения) |
| `geo` | GEOGRAPHY(Point,4326) | нет | — | координаты, гео-поиск «рядом со мной» |
| `location_landmarks` | TEXT | да | `NULL` | текстовый ориентир («рядом с рынком») — дополнение к GPS, не замена (FR-06) |
| `phone` | TEXT | да | `NULL` | контактный телефон учреждения |
| `email` | TEXT | да | `NULL` | контактный email учреждения |
| `website` | TEXT | да | `NULL` | — |
| `socials` | JSONB | да | `NULL` | `{instagram?, telegram?, facebook?}` |
| `cover_photo_s3_key` | TEXT | да | `NULL` | обложка профиля, S3-ключ |
| `age_range` | TEXT | да | `NULL` | текстовый возрастной диапазон («1-3 года») |
| `tag` | JSONB | да | `NULL` | `{ru,tg}` — короткий бейдж на карточке в каталоге |
| `license_no` | TEXT | да | `NULL` | свободный текст, проверяется модератором вручную |
| `languages` | TEXT[] | да | `NULL` | языки обучения |
| `program_level` | TEXT[] | да | `NULL` | ступень/программа (FR-28): для школ начальная/основная/средняя, для вузов бакалавриат/магистратура/докторантура |
| `curriculum` | TEXT[] | да | `NULL` | `state`\|`bilingual`\|`international`\|`stem` |
| `price` | INT | да | `NULL` | примерная стоимость |
| `transport_type` | TEXT | да | `NULL` | `own_bus`\|`minibus`\|`taxi`\|`parent_coop`\|`none` |
| `transport_cost` | INT | да | `NULL` | — |
| `transport_areas` | TEXT[] | да | `NULL` | охватываемые районы |
| `meals_available` | TEXT | да | `NULL` | `yes`\|`no`\|`bring_own` |
| `meals_type` | TEXT | да | `NULL` | `hot`\|`breakfast`\|`buffet`\|`none` |
| `meals_cost` | INT | да | `NULL` | — |
| `meals_halal` | BOOL | да | `NULL` | — |
| `discount_available` | BOOL | нет | `false` | FR-26 |
| `discount_type` | TEXT[] | да | `NULL` | `needs_based`\|`merit_or_competition`\|`large_family`\|`other` |
| `discount_details` | TEXT | да | `NULL` | разовое текстовое поле |
| `verified` | BOOL | нет | `false` | верификация владельца профиля (FR-34) — НЕ то же самое, что `moderation_status` |
| `moderation_status` | TEXT | нет | `'pending'` | `pending`\|`approved`\|`rejected` (CHECK) — публикация в каталоге |
| `plan` | TEXT | нет | `'free'` | `free`\|`pro`\|`enterprise` — влияет только на бизнес-лимиты кабинета учреждения (см. `docs/EduHub_Pricing_Tiers.md`), никогда на поиск/сортировку/`rating_avg` (обновлено 2026-08-25 — буст позиции полностью убран как продуктовое решение) |
| `plan_expires_at` | TIMESTAMPTZ | да | `NULL` | истечение тарифа автоматически возвращает лимиты на Free |
| `founded` | INT | да | `NULL` | год основания |
| `students_count` | INT | да | `NULL` | — |
| `rating_avg` | NUMERIC(3,2) | да | `NULL` | денормализация — `NULL` означает «недостаточно отзывов» (FR-29), не «ноль»; пишет только `catalog` через порт `RatingSync`, источник расчёта — веха 4 |
| `review_count` | INT | нет | `0` | денормализация, синхронно с `rating_avg` |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | инкрементируется при любой правке — основа ETag (задача 28) и оптимистической блокировки (E3.4) |

**Связи:** 1—N `catalog.institution_owners`, `institution_staff`, `institution_gallery`, `institution_alumni`, `news_articles`, `institution_metrics`; 1—N `auth.children` (принимающая сторона кросс-схемного FK); referenced by `reviews.reviews`, `communications.messages`, `communications.vacancies`.
**Индексы:** `GIST(geo)`; `GIN(name jsonb_path_ops)`; `GIN((name->>'ru') gin_trgm_ops)`, `GIN((name->>'tg') gin_trgm_ops)` (подстрочный поиск на обоих языках — trigram, не tsvector, см. план веха 1 задача 18); `GIN(types)`; `GIN(curriculum)`; `GIN(program_level)`; `btree(region, district)`; частичный `btree(rating_avg DESC, id) WHERE moderation_status='approved'`; частичный `btree(price) WHERE moderation_status='approved'`; частичный `UNIQUE(lower(name->>'ru'), region, district) WHERE moderation_status <> 'rejected'` (защита от двойного сабмита формы регистрации — см. E3.3 в плане; `rejected`-заявки не блокируют повторную попытку с исправленными данными).

### `catalog.institution_metrics`

Денормализация 8 метрик рейтинга (не только общего `rating_avg`) — карточка учреждения (`GetByID`, задача 24 плана) собирается одним SQL-запросом и должна отдавать разбивку по метрикам, а `catalog` не вправе читать чужую схему `reviews` напрямую. Пишет только `catalog` через расширенный порт `RatingSync.Apply(ctx, instID, metrics []MetricScore)`, источник расчёта — веха 4 (`reviews.institution_rating_agg`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | часть составного PK |
| `metric_key` | TEXT | нет | — | часть составного PK, те же 8 значений что в `reviews.review_metrics` |
| `weighted_avg` | NUMERIC(3,2) | да | `NULL` | `NULL` — недостаточно отзывов по этой метрике |
| `review_count` | INT | нет | `0` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions` (внутри своей схемы — FK разрешён).
**Индексы:** `PK(institution_id, metric_key)`.

### `catalog.institution_owner_verifications`

Материалы верификации владельца профиля учреждения (FR-34) — сейчас `institutions.verified BOOL` хранил только результат, без документа/обоснования, на основе которых модератор принял решение (добавлено 2026-08-25). По закону РТ о лицензировании — **государственные детсады и школы (начальное/основное/общее среднее) лицензированию не подлежат вообще**, лицензия нужна только частным организациям (лицеи/гимназии/колледжи/вузы/частные центры) — `document_type` отражает эту разницу, а не считает отсутствие лицензии у госучреждения дефектом.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | каждая попытка — отдельная строка, история не удаляется даже при отклонении |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `submitted_by` | UUID | нет | — | кто подал заявку — по значению |
| `document_type` | TEXT | нет | — | `license`\|`state_status_confirmation`\|`appointment_proof`\|`business_registration`\|`manual_exception`\|`other` (CHECK) |
| `document_s3_key` | TEXT | да | `NULL` | пусто при `manual_exception` |
| `license_no_claimed` | TEXT | да | `NULL` | заявленный номер лицензии на момент подачи — снимок, `institutions.license_no` может измениться позже |
| `verification_notes` | TEXT | да | `NULL` | **обязательно** (проверка в usecase) при `document_type='manual_exception'` — письменное обоснование модератора, почему подтвердил без документа |
| `status` | TEXT | нет | `'pending'` | `pending`\|`approved`\|`rejected` (CHECK) |
| `reviewed_by` | UUID | да | `NULL` | модератор — по значению |
| `reviewed_at` | TIMESTAMPTZ | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions` (внутри своей схемы — FK разрешён).
**Индексы:** `btree(institution_id, created_at DESC)`; `btree(status)` (очередь модератора).
**Эффект при `approved`:** `catalog.institutions.verified` ставится в `true`; предыдущие отклонённые попытки не удаляются — полная история заявок хранится.

### `catalog.institution_owners`

Кто из пользователей владеет профилем учреждения (для RBAC на мутациях, E3.3-E3.4).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `user_id` | UUID | нет | — | ссылка на `auth.users.id` по значению, без физического FK (правило: кросс-схемно — только `user_id`, без FK) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions`.
**Индексы:** `UNIQUE(institution_id, user_id)`; `btree(user_id)`.

### `catalog.institution_staff`

Педагоги/персонал учреждения (`Person` во фронте, `web/lib/data.ts:157-173`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `name` | JSONB | нет | — | `{ru,tg}` |
| `role_type` | TEXT | нет | — | `director`\|`teacher`\|`coach`\|`psychologist`\|`admin` |
| `role_label` | JSONB | нет | — | `{ru,tg}` — человекочитаемое название должности |
| `subject` | JSONB | да | `NULL` | `{ru,tg}` — предмет, если применимо |
| `photo_url` | TEXT | да | `NULL` | S3-ключ |
| `exp` | TEXT | да | `NULL` | стаж |
| `bio` | JSONB | да | `NULL` | `{ru,tg}` |
| `education` | JSONB | да | `NULL` | массив `{ru,tg}` записей |
| `email` | TEXT | да | `NULL` | — |
| `phone` | TEXT | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions`; принимающая сторона для `catalog.achievements` (полиморфно, без FK).
**Индексы:** `btree(institution_id)`.

### `catalog.achievements`

Достижения учреждения/сотрудника/ученика — полиморфная связь (`owner_type`+`owner_id`), поэтому **без FK** (Postgres не поддерживает FK на «один из нескольких родителей»; целостность — на уровне usecase).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `owner_type` | TEXT | нет | — | `institution`\|`staff`\|`student` (CHECK) |
| `owner_id` | UUID | нет | — | id владельца в соответствующей таблице — без FK |
| `title` | JSONB | нет | — | `{ru,tg}` |
| `year` | INT | нет | — | — |
| `category` | TEXT | нет | — | `gold`\|`silver`\|`bronze`\|`special` |
| `description` | JSONB | нет | — | `{ru,tg}` |
| `links` | JSONB | да | `NULL` | массив `{label,url}` |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет FK (полиморфная связь) — целостность проверяется в usecase при записи.
**Индексы:** `btree(owner_type, owner_id)`.

### `catalog.institution_gallery`

Фото/видео учреждения.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `s3_key` | TEXT | нет | — | ключ в объектном хранилище, не сам файл |
| `label` | JSONB | да | `NULL` | `{ru,tg}` подпись |
| `sort_order` | INT | нет | `0` | порядок для перетаскивания в кабинете (FR-07) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions`.
**Индексы:** `btree(institution_id, sort_order)`.

### `catalog.institution_alumni`

Выпускники (для школ/вузов).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `name` | JSONB | нет | — | `{ru,tg}` |
| `photo_url` | TEXT | да | `NULL` | S3-ключ |
| `grad_year` | INT | нет | — | — |
| `now_label` | JSONB | да | `NULL` | `{ru,tg}` — «сейчас: ...» |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions`.
**Индексы:** `btree(institution_id)`.

### `catalog.news_articles`

Новости учреждения.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID FK→`catalog.institutions(id)` ON DELETE CASCADE | нет | — | — |
| `title` | JSONB | нет | — | `{ru,tg}` |
| `category` | JSONB | да | `NULL` | `{ru,tg}` |
| `cover_s3_key` | TEXT | да | `NULL` | — |
| `video_url` | TEXT | да | `NULL` | внешняя ссылка, без транскодирования (FR-07 принцип) |
| `content` | JSONB | нет | — | `{ru,tg}` |
| `tags` | JSONB | да | `NULL` | массив `{ru,tg}` |
| `status` | TEXT | нет | `'draft'` | `published`\|`draft` |
| `views_count` | INT | нет | `0` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `catalog.institutions`.
**Индексы:** `btree(institution_id, status, created_at DESC)`.

---

## Схема `reviews`

Владелец: сервис Reviews/Ratings (веха 4).

### `reviews.reviews`

Отзыв с верификацией через `Child`.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID | нет | — | автор, по значению (кросс-схемно без FK) |
| `institution_id` | UUID | нет | — | по значению (кросс-схемно без FK) — существование проверяется в usecase |
| `child_id` | UUID | нет | — | по значению — привязка, дающая право на отзыв (FR-15), проверяется через порт `auth.ChildLinkExists` |
| `text` | TEXT | нет | — | свободный текст на любом языке, независимо от языка интерфейса платформы (обновлено 2026-08-25 — было `JSONB{ru,tg}`; отзыв пишет живой человек в моменте на одном языке, не переводит сам себя на оба; тот же принцип, что уже применён к `communications.messages.body`) |
| `reply` | TEXT | да | `NULL` | ответ учреждения (FR-17), тот же принцип |
| `status` | TEXT | нет | `'pending'` | `pending`\|`approved`\|`rejected`\|`disputed`\|`resolved_kept`\|`resolved_removed` (обновлено 2026-08-25 — `disputed` больше не конечное состояние, спор разрешается в одно из двух явных состояний) |
| `verified_at_publish` | BOOL | нет | — | **снимок** факта верификации на момент публикации (был ли `auth.children.confirmation_status='confirmed'` в этот момент) — критерий разрешения спора (FR-35: «была ли подтверждённая привязка **на момент публикации**») смотрит на этот снимок, не на текущее состояние `auth.children`, которое могло измениться (родитель удалил привязку, учреждение отозвало подтверждение) |
| `dispute_deadline` | TIMESTAMPTZ | да | `NULL` | SLA 72ч на разрешение спора (FR-35) |
| `disputed_by` | UUID | да | `NULL` | кто эскалировал спор (обычно представитель учреждения) — по значению |
| `dispute_reason` | TEXT | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | используется в decay-формуле (`age_days`) |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** 1—N `reviews.review_metrics`; логически N—1 `auth.users`, N—1 `catalog.institutions`, N—1 `auth.children` (все три — без физического FK, кросс-схемные, проверяются в usecase/через порты).
**Индексы:** `UNIQUE(user_id, institution_id, child_id)` — идемпотентность отзыва на уровне БД; `btree(institution_id, status)`.

### `reviews.review_metrics`

8 метрик отзыва, нормализовано по строкам (не массив — нужно для decay-агрегации по времени и по метрике одновременно).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `review_id` | UUID FK→`reviews.reviews(id)` ON DELETE CASCADE | нет | — | — |
| `metric_key` | TEXT | нет | — | `quality`\|`conditions`\|`safety`\|`food`\|`transport`\|`price`\|`parent_involvement`\|`inclusivity` (CHECK, 8 значений — см. design decisions log в плане) |
| `score` | SMALLINT | нет | — | `1..5` (CHECK) |

**Связи:** N—1 `reviews.reviews`.
**Индексы:** `UNIQUE(review_id, metric_key)`.

### `reviews.institution_rating_agg`

Денормализованный агрегат по метрике — источник для синхронизации в `catalog.institutions` (порт `RatingSync`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `institution_id` | UUID | нет | — | по значению, часть составного PK |
| `metric_key` | TEXT | нет | — | часть составного PK |
| `weighted_avg` | NUMERIC(3,2) | нет | — | `Σ(w_i·score_i) / Σ(w_i)`, decay-формула (E4.3) |
| `review_count` | INT | нет | `0` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** логически N—1 `catalog.institutions` (без FK).
**Индексы:** `PK(institution_id, metric_key)`.

### `reviews.employer_reviews`

Отзыв сотрудника (текущего/бывшего) об учреждении **как о работодателе** — добавлено 2026-08-25, ранее сознательно отложено (7-я волна фронта: ломало модель верификации через `Child`, метрики несовместимы с родительскими, риск для монетизации). Все три причины сняты дизайном: верификация через модератора (не через `Child`), полностью отдельные метрики (не влияют на `catalog.institution_metrics`/`rating_avg`), видимость ограничена соискателями (см. ниже) — учреждение как платящий клиент никогда не видит это на своей публичной родительской витрине.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID | нет | — | по значению |
| `institution_id` | UUID | нет | — | по значению |
| `employment_claim_id` | UUID | нет | — | по значению — ссылка на подтверждённую (`verification_status='verified'`) запись `auth.employment_claims` |
| `text` | TEXT | нет | — | свободный текст (НЕ `JSONB{ru,tg}` — пишет один человек на одном языке) |
| `reply` | TEXT | да | `NULL` | ответ учреждения — разрешён (не может скрыть отзыв, но может публично ответить — снимает конфликт интересов иначе) |
| `status` | TEXT | нет | `'pending'` | `pending`\|`approved`\|`rejected` — тот же модерационный цикл, что и родительские отзывы (FR-16) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `auth.employment_claims` (по значению — `reviews` не владеет схемой `auth`).
**Индексы:** `UNIQUE(user_id, institution_id)` — один отзыв о работодателе на пару; `btree(institution_id, status)`.
**Видимость (жёстко ограничена, не просто RBAC-роль):** `GET /api/v1/institutions/{id}/employer-reviews` требует аутентификации И наличия собственной записи в `communications.applicants` (то есть пользователь — соискатель). Родители, гости, обычные пользователи без анкеты соискателя не видят ни через профиль учреждения, ни по прямой ссылке.

### `reviews.employer_review_metrics`

5 метрик отзыва о работодателе, нормализовано по строкам (тот же паттерн, что `review_metrics`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `employer_review_id` | UUID FK→`reviews.employer_reviews(id)` ON DELETE CASCADE | нет | — | — |
| `metric_key` | TEXT | нет | — | `salary_conditions`\|`management`\|`team_atmosphere`\|`workload`\|`professional_growth` (CHECK, 5 значений) |
| `score` | SMALLINT | нет | — | `1..5` (CHECK) |

**Связи:** N—1 `reviews.employer_reviews`.
**Индексы:** `UNIQUE(employer_review_id, metric_key)`.

### `reviews.outbox`

Transactional outbox — публикация отзыва пишет отзыв+метрики+outbox-запись одной транзакцией; отдельный релей публикует в Redis Stream.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | он же `event_id` для идемпотентности консьюмера |
| `topic` | TEXT | нет | — | например `review.approved` |
| `payload` | JSONB | нет | — | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `published_at` | TIMESTAMPTZ | да | `NULL` | `NULL` = ещё не отправлено, релей выбирает `FOR UPDATE SKIP LOCKED` |

**Связи:** нет FK (событийная таблица).
**Индексы:** частичный `btree(created_at) WHERE published_at IS NULL`.

---

## Схема `moderation`

Владелец: сквозной пакет `internal/moderation`, используется всеми сервисами через порт `Recorder`. Вынесен в отдельную схему специально, чтобы `catalog`/`reviews`/`communications` не писали в чужие схемы при модерации своих сущностей.

### `moderation.audit_log`

Неизменяемый (append-only) журнал всех решений модерации.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `actor_id` | UUID | нет | — | кто принял решение, по значению |
| `actor_role` | TEXT | нет | — | снимок роли на момент действия (роль могла измениться позже) |
| `action` | TEXT | нет | — | `institution.approve`\|`institution.reject`\|`review.approve`\|`review.reject`\|`review.dispute_resolve`\|`owner.verify`\|... |
| `target_type` | TEXT | нет | — | `institution`\|`review`\|`vacancy`\|... |
| `target_id` | UUID | нет | — | — |
| `reason_code` | TEXT | да | `NULL` | структурированная причина: `no_license`\|`duplicate`\|`fake_suspected`\|`other` — обязателен для reject-действий (проверка в usecase) |
| `reason_text` | TEXT | да | `NULL` | свободный текст в дополнение к коду |
| `payload_diff` | JSONB | да | `NULL` | что изменилось |
| `request_id` | TEXT | нет | — | для трассировки |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет FK — журнал переживает удаление своих акторов/целей.
**Индексы:** `btree(target_type, target_id, created_at DESC)`; `btree(actor_id)`.
**Инвариант:** запись создаётся только через `Recorder.Record(ctx, tx, Entry)`, принимающий `pgx.Tx` — физически невозможно записать вне транзакции самого изменения (см. план, риск №7).

### `moderation.queue_items`

Очередь модерации — одна общая на всех модераторов (MVP-решение), с claim-паттерном и приоритетом по флагам.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `target_type` | TEXT | нет | — | `institution`\|`review`\|`owner_verification` |
| `target_id` | UUID | нет | — | — |
| `status` | TEXT | нет | `'pending'` | `pending`\|`claimed`\|`resolved` |
| `priority` | SMALLINT | нет | `0` | выше = раньше в очереди; поднимается анти-спам детектором (FR-18) |
| `flagged_reason` | TEXT | да | `NULL` | почему приоритет поднят (например `velocity_anomaly`) |
| `claimed_by` | UUID | да | `NULL` | модератор, взявший в работу — по значению |
| `claimed_at` | TIMESTAMPTZ | да | `NULL` | claim истекает через 15 минут (обрабатывается фоново, не хранится отдельным полем) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет FK на `target_id` (полиморфно, как `achievements`).
**Индексы:** частичный `btree(priority DESC, created_at) WHERE status='pending'`; `btree(claimed_by, claimed_at) WHERE status='claimed'` (для таймаут-джоба возврата в пул).

---

## Схема `communications`

Владелец: сервис Communications (веха 5). Чат, уведомления, вакансии и вся HR-часть MVP.

### `communications.conversations`

Диалог (FR-19) — **обобщено 2026-08-25**: не только «родитель ↔ учреждение», но и «родитель ↔ родитель» (пользователи тоже могут общаться друг с другом). Оба участника — полиморфные слоты, каждый может быть либо пользователем, либо учреждением. Раздельные `*_last_read_at` по сторонам, поскольку прочтение одной стороной не означает прочтение другой.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | он же `conv` в `chat:conv:{id}` (Redis Pub/Sub канал, план E5.2) |
| `participant_a_type` | TEXT | нет | — | `user`\|`institution` (CHECK) |
| `participant_a_id` | UUID | нет | — | по значению |
| `participant_b_type` | TEXT | нет | — | `user`\|`institution` (CHECK) |
| `participant_b_id` | UUID | нет | — | по значению |
| `participant_a_last_read_at` | TIMESTAMPTZ | да | `NULL` | — |
| `participant_b_last_read_at` | TIMESTAMPTZ | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** 1—N `communications.messages`.
**Индексы:** `UNIQUE(participant_a_type, participant_a_id, participant_b_type, participant_b_id)` — при создании диалога usecase **канонизирует порядок** участников (например сортировкой по `type`, затем по `id`) до вставки, чтобы пара «X↔Y» и «Y↔X» не создали два разных диалога — канонизация в usecase, не в БД (тот же принцип, что уже применён для нескольких других полей сегодня).
**Список диалогов пользователя** (`GET /api/v1/conversations`) — `WHERE (participant_a_type=$my_type AND participant_a_id=$my_id) OR (participant_b_type=$my_type AND participant_b_id=$my_id)`, каждая запись отдаёт «собеседника» как `{type:"user", display_name}` или `{type:"institution", name}` — фронт рендерит по-разному, схема одна.
**Открытый вопрос (не блокирует схему):** механизм **обнаружения** — как пользователь находит `user_id` другого родителя, чтобы начать диалог (приватность: не должен раскрывать, что конкретный человек — родитель конкретного учреждения без его согласия) — отдельная продуктовая задача до реализации вехи 5.

### `communications.messages`

Отдельные реплики внутри диалога — кто конкретно написал, различимо по `sender_type`/`sender_id` (учреждение может иметь несколько представителей с доступом к чату).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | клиент использует его как `message_id` для дедупликации доставки по WS |
| `conversation_id` | UUID FK→`communications.conversations(id)` ON DELETE CASCADE | нет | — | — |
| `sender_type` | TEXT | нет | — | `user`\|`institution` (CHECK) |
| `sender_id` | UUID | нет | — | конкретный `user_id` (родитель или представитель учреждения) — по значению |
| `body` | TEXT | нет | — | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `communications.conversations` (внутри своей схемы — FK разрешён).
**Индексы:** `btree(conversation_id, created_at)`.
**Rate-limit:** частота отправки — по `sender_id` (Redis-счётчик), не в схеме — защита от спама в чат.

### `communications.visit_requests`

Заявка на визит в учреждение (уже есть на фронте, `web/app/(site)/institutions/[id]/page.tsx:344-352` — форма собирает имя/телефон/дату; `.claude/rules/architecture.md` прямо называет её среди мутаций, требующих идемпотентности).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID | нет | — | по значению |
| `user_id` | UUID | да | `NULL` | гость может оставить заявку без регистрации |
| `name` | TEXT | нет | — | ФИО заявителя (PII — собирается с явной целью, срок хранения ограничен) |
| `phone` | TEXT | нет | — | — |
| `preferred_date` | DATE | да | `NULL` | — |
| `status` | TEXT | нет | `'new'` | `new`\|`contacted`\|`closed` (CHECK) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет физических FK (кросс-схемно).
**Индексы:** `btree(institution_id, status, created_at)`; частичный `UNIQUE(institution_id, phone) WHERE status='new'` — защита от дубль-клика (идемпотентность, повторная заявка на ту же институцию с тем же телефоном пока предыдущая ещё не обработана — не создаёт вторую запись).
**Rate-limit:** публичный эндпоинт, собирает PII от гостя без аутентификации — обязательный rate limit по IP+телефону в usecase.

### `communications.notifications`

In-app уведомления (FR-20 — часть общего уведомительного механизма, push/SMS — вне БД, через `notify.Sender`).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID | нет | — | по значению |
| `type` | TEXT | нет | — | `message`\|`review_reply`\|`news`\|`achievement` |
| `text` | TEXT | нет | — | — |
| `link` | TEXT | да | `NULL` | — |
| `read` | BOOL | нет | `false` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет физических FK.
**Индексы:** `btree(user_id, read, created_at DESC)`.

### `communications.push_subscriptions`

Web Push подписки (Push API + Service Worker, встроенный браузерный стандарт — НЕ внешний платный провайдер, в отличие от SMS/email; добавлено 2026-08-25 в MVP-скоуп, т.к. NFR CLAUDE.md называет push основным каналом PWA и здесь нет причины откладывать вместе с SMS/email). Один пользователь может иметь несколько подписок (несколько устройств/браузеров).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID | нет | — | по значению |
| `endpoint` | TEXT | нет | — | URL пуш-сервера браузера пользователя (Google/Mozilla/Apple — у каждого свой), уникален по своей природе |
| `p256dh_key` | TEXT | нет | — | ключ шифрования сообщения (Web Push API) |
| `auth_key` | TEXT | нет | — | ключ аутентификации (Web Push API) |
| `user_agent` | TEXT | да | `NULL` | для отладки/отображения «какое устройство» пользователю |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет физических FK (кросс-схемно).
**Индексы:** `UNIQUE(endpoint)`.

### `communications.vacancies`

Вакансия учреждения (FR-36).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID | нет | — | по значению |
| `title` | JSONB | нет | — | `{ru,tg}` |
| `description` | JSONB | нет | — | `{ru,tg}` |
| `requirements` | JSONB | да | `NULL` | массив `{ru,tg}` |
| `salary_from` | INT | да | `NULL` | — |
| `salary_to` | INT | да | `NULL` | — |
| `employment` | JSONB | нет | — | `{ru,tg}` |
| `status` | TEXT | нет | `'draft'` | `published`\|`draft` |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** нет физического FK на `catalog.institutions`; 1—N `communications.applications` (по значению).
**Индексы:** `btree(institution_id, status)`.

### `communications.applicants`

Профиль соискателя (резюме).

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `user_id` | UUID | нет | — | по значению, один профиль на пользователя |
| `name` | JSONB | нет | — | `{ru,tg}` |
| `photo_url` | TEXT | да | `NULL` | S3-ключ |
| `position` | JSONB | нет | — | желаемая должность `{ru,tg}` |
| `bio` | JSONB | да | `NULL` | `{ru,tg}` |
| `education` | JSONB | да | `NULL` | массив `{ru,tg}` |
| `experience` | JSONB | да | `NULL` | массив `{ru,tg}` |
| `skills` | JSONB | да | `NULL` | массив `{ru,tg}` |
| `email` | TEXT | да | `NULL` | скрывается по `hide_contacts` |
| `phone` | TEXT | да | `NULL` | скрывается по `hide_contacts` |
| `cv_s3_key` | TEXT | да | `NULL` | — |
| `visibility` | TEXT | нет | `'draft'` | `draft`\|`on_response`\|`public` |
| `hide_contacts` | BOOL | нет | `true` | скрывать контакты до ответа в чате |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |
| `updated_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** `UNIQUE(user_id)`; 1—N `communications.applications`, `communications.employer_responses`, `communications.applicant_achievements`.
**Индексы:** `UNIQUE(user_id)`; частичный `btree(visibility) WHERE visibility='public'` (публичный список `/applicants`).
**Важно (PII):** контакты фильтруются на уровне SQL-проекции при `hide_contacts=true` и `visibility≠'public'`, не в DTO-маппинге после чтения (см. E5.5 в плане).

### `communications.applicant_achievements`

Достижения соискателя (`web/lib/data.ts:2002`, `Applicant.achievements`) — добавлено 2026-08-25. **Не переиспользует** полиморфную `catalog.achievements` (та в схеме `catalog`, `communications` не может писать в чужую таблицу — правило владения схемами); структура продублирована, тот же паттерн, что уже применён для `reviews.employer_review_metrics` отдельно от `reviews.review_metrics`.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `applicant_id` | UUID FK→`communications.applicants(id)` ON DELETE CASCADE | нет | — | — |
| `title` | TEXT | нет | — | свободный текст (не `JSONB{ru,tg}` — пишет сам соискатель, тот же принцип что находка #11) |
| `year` | INT | да | `NULL` | — |
| `category` | TEXT | да | `NULL` | `gold`\|`silver`\|`bronze`\|`special` (CHECK), необязательно |
| `description` | TEXT | да | `NULL` | — |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `communications.applicants` (внутри своей схемы — FK разрешён).
**Индексы:** `btree(applicant_id)`.

### `communications.applications`

Отклик соискателя на вакансию — идемпотентно на уровне БД.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `applicant_id` | UUID FK→`communications.applicants(id)` ON DELETE CASCADE | нет | — | — |
| `vacancy_id` | UUID FK→`communications.vacancies(id)` ON DELETE CASCADE | нет | — | — |
| `status` | TEXT | нет | `'sent'` | `sent`\|`viewed`\|`closed` |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `communications.applicants`, N—1 `communications.vacancies` (обе внутри своей схемы — FK разрешены).
**Индексы:** `UNIQUE(applicant_id, vacancy_id)` — идемпотентность отклика.

### `communications.employer_responses`

Обращение учреждения к соискателю напрямую (обратное направление к `applications`) — идемпотентно.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID PK | нет | `gen_random_uuid()` | — |
| `institution_id` | UUID | нет | — | по значению |
| `applicant_id` | UUID FK→`communications.applicants(id)` ON DELETE CASCADE | нет | — | — |
| `message` | TEXT | нет | — | свободный текст (обновлено 2026-08-25 — было `JSONB{ru,tg}`, пишет живой человек, тот же принцип что `reviews.reviews.text`) |
| `created_at` | TIMESTAMPTZ | нет | `now()` | — |

**Связи:** N—1 `communications.applicants`.
**Индексы:** `UNIQUE(institution_id, applicant_id)` — идемпотентность обращения; защита от спама (FR-37) — дополнительный rate limit в usecase, не в схеме.

---

## Схема `analytics`

Владелец: сервис Analytics (веха 6).

### `analytics.profile_events`

Сырые события просмотра/клика. Партиционирована по месяцам (`PARTITION BY RANGE(occurred_at)`) — retention 90 дней снимается через `DROP PARTITION`, не `DELETE`.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `id` | UUID | нет | `gen_random_uuid()` | часть композитного PK с `occurred_at` (партиционирование требует включения ключа партиции в PK) |
| `institution_id` | UUID | нет | — | по значению |
| `event_type` | TEXT | нет | — | `view`\|`click_phone`\|`click_website`\|`click_chat`\|... |
| `occurred_at` | TIMESTAMPTZ | нет | `now()` | ключ партиции |
| `session_hash` | TEXT | нет | — | хешированный идентификатор сессии, не PII |
| `ip_hash` | TEXT | нет | — | **хешированный/усечённый IP, не сырой** (PII-минимизация) |
| `user_agent_class` | TEXT | нет | — | `bot`\|`mobile`\|`desktop` — классифицированное значение, не полная UA-строка |
| `referrer_class` | TEXT | нет | — | `search`\|`catalog`\|`map`\|`direct`\|`external` (CHECK, добавлено 2026-08-25, FR-22/23 «источники трафика») — классифицируется на бэкенде из `Referer`-заголовка при записи события, сырой URL не сохраняется (тот же принцип, что `user_agent_class`: полный `Referer` может содержать PII в query-параметрах источника) |

**Связи:** нет FK (событийная таблица, высокий объём записи).
**Индексы:** `btree(institution_id, occurred_at)` на каждой партиции.
**Путь записи:** асинхронный, через bounded-канал + batch `COPY` — переполнение канала отбрасывает событие (lossy by design, см. план E6.2), не блокирует основной HTTP-запрос.

### `analytics.profile_events_daily`

Агрегат по дням — материализуется ночным джобом, дашборд читает только отсюда.

| Поле | Тип | Nullable | Default | Описание |
|---|---|---|---|---|
| `institution_id` | UUID | нет | — | часть PK |
| `day` | DATE | нет | — | часть PK |
| `event_type` | TEXT | нет | — | часть PK |
| `count` | INT | нет | `0` | — |

**Связи:** нет FK.
**Индексы:** `PK(institution_id, day, event_type)` — упрощает идемпотентный `INSERT ... ON CONFLICT DO UPDATE` (E6.3).

---

## Сводка: все таблицы по схемам

| Схема | Таблицы | Кол-во |
|---|---|---|
| `platform` | `idempotency_keys`, `seed_refs` | 2 |
| `auth` | `users`, `refresh_tokens`, `children`, `oauth_identities`, `verification_codes`, `employment_claims` | 6 |
| `catalog` | `institutions`, `institution_owners`, `institution_staff`, `achievements`, `institution_gallery`, `institution_alumni`, `news_articles`, `institution_metrics`, `institution_owner_verifications` | 9 |
| `reviews` | `reviews`, `review_metrics`, `institution_rating_agg`, `outbox`, `employer_reviews`, `employer_review_metrics` | 6 |
| `moderation` | `audit_log`, `queue_items` | 2 |
| `communications` | `conversations`, `messages`, `notifications`, `push_subscriptions`, `vacancies`, `applicants`, `applicant_achievements`, `applications`, `employer_responses`, `visit_requests` | 10 |
| `analytics` | `profile_events`, `profile_events_daily` | 2 |
| **Итого** | | **37** |

## Диаграмма связей (текстовая, ключевые FK)

```
auth.users ──┬─< auth.refresh_tokens
             ├─< auth.oauth_identities
             ├─< auth.verification_codes
             ├─< auth.children >── FK ──> catalog.institutions   (согласованное исключение)
             ├─< auth.employment_claims >── FK ──> catalog.institutions   (тот же паттерн, верификация модератором)
             └─  (по user_id, без FK) владелец/актор во всех остальных схемах

catalog.institutions ──┬─< catalog.institution_owners
                        ├─< catalog.institution_staff ──< catalog.achievements (полиморфно, без FK)
                        ├─< catalog.institution_gallery
                        ├─< catalog.institution_alumni
                        ├─< catalog.news_articles
                        ├─< catalog.institution_metrics
                        ├─  (по institution_id, без FK) reviews.reviews
                        ├─  (по participant_*_id, без FK) communications.conversations
                        ├─  (по institution_id, без FK) communications.visit_requests
                        └─  (по institution_id, без FK) communications.vacancies

reviews.reviews ──┬─< reviews.review_metrics
                   └─  (агрегируется в) reviews.institution_rating_agg ──> порт RatingSync (per-metric) ──> catalog.institution_metrics + catalog.institutions.rating_avg

reviews.employer_reviews ──< reviews.employer_review_metrics   (полностью отдельно от родительского рейтинга; видимость — только соискателям, см. описание таблицы)

communications.conversations (participant_a/b — user ИЛИ institution, полиморфно) ──< communications.messages (sender_type различает участников)

communications.applicants ──┬─< communications.applications >── FK ──> communications.vacancies
                             └─< communications.employer_responses

moderation.audit_log, moderation.queue_items — полиморфные target_type/target_id, без FK, пишутся через порт Recorder всеми сервисами
```

**Релевантные файлы:** полный план реализации и обоснование решений — `docs/EduHub_Backend_Development_Plan.md`. Контракты фронта для сверки полей — `web/lib/data.ts`, `web/lib/app-state.tsx`.
