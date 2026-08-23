# EduHub — План разработки платформы (Product & Backend), v1.1

**Дата:** 2026-08-22
**Источники:** `EduHub_Technical_Specification.docx` (SRS v2.1), `EduHub_Business_Plan_Pro.docx` (v3.3), полный аудит фактического состояния `web/` (Next.js-прототип, 26 роутов, mock-данные, 13+ волн правок).
**Статус:** рабочий документ синхронизации SRS↔факт + модель данных для Go/Postgres-бэкенда (SRS §5, ещё не начат). Не заменяет SRS/бизнес-план как юридический источник правды — но фиксирует решения, которые в них ещё не отражены.

Метод: (1) построчное чтение SRS v2.1 и бизнес-плана v3.3 из `.docx`, (2) полный аудит TypeScript-сущностей `lib/data.ts`/`lib/app-state.tsx`, (3) каталог всех 26 `page.tsx`/`layout.tsx`/top-level компонентов, (4) сверка расхождений и решения пользователя по 4 ключевым развилкам (см. §0).

---

## 0. Решения по развилкам (приняты пользователем 2026-08-22)

| # | Развилка | Решение |
|---|---|---|
| 1 | FR-37 (обратный поиск CV — `EmployerResponse`, учреждение пишет соискателю напрямую через `/applicants`) — уже реализован на фронте (12-я волна), но SRS §11 явно относит «CV-база, обратный поиск» ко второй волне, и консультация `eduhub-advisor` (7-я волна, п.9.4) отметила это как незакрытый конфликт | **Оставить в MVP**, формализовать как **FR-37**. Осознанное отступление от SRS-текста и прежней рекомендации advisor'а — принято явно, не по умолчанию |
| 2 | FR-04 (сравнение 2-3 учреждений) — было в исходном SRS MVP, убрано с фронта во 2-й волне по прямому запросу пользователя | **Официально исключить из MVP** в этом документе, перенести в Core/вторую волну |
| 3 | `meals_halal`/FR-13 «халяль» — есть в SRS/бизнес-плане, но halal убран с фронта «полностью и everywhere» ранним явным решением | **Синхронизировать документы под код**: halal убирается из модели данных и FR-13. Питание — 3 поля (наличие/тип/стоимость), без халяль |
| 4 | Роль «Администратор МОН» (SRS §3: реестр/верификация/аналитика) vs фактическая роль `/admin` в коде (тарифы, maintenance-режим) — разные по смыслу | **Слить в одну роль.** `/admin` в бэкенд-плане = «Администратор МОН» с совмещёнными правами: реестр и верификация учреждений, агрегированная аналитика (FR-24) **плюс** тарифы Pro/Enterprise и maintenance-режим платформы |

Не спрашивалось отдельно (низкий риск, дефолт с обоснованием): актуализация самих `.docx` (SRS v2.2 / бизнес-план v3.4) — **отложена**, этот план временно служит единым журналом синхронизации. Если проект перейдёт к найму бэкенд-команды — рекомендуется всё же выпустить v2.2/v3.4 по образцу Приложения A, чтобы .docx не разошлись с этим планом ещё сильнее.

---

## 1. Роли и RBAC (сведено, финально)

RBAC-роль (проверяется на каждом API-эндпоинте бэкенда) и **факт профиля** (что пользователь имеет/умеет) — разные вещи. Это архитектурное решение уже принято в коде (`lib/app-state.tsx`, «модель B», 7-я волна) и сохраняется здесь.

### 1.1 RBAC-роли (5)

| Роль | Права | Кто | Этап |
|---|---|---|---|
| Гость | Просмотр каталога, профилей, отзывов, поиск | Не авторизован | MVP |
| Пользователь (`user`) | Всё гостя + отзывы\*, чат, избранное, дети, резюме соискателя, отклики на вакансии, обращения к учреждениям | Родитель / студент / соискатель — это факты профиля, не суб-роли | MVP |
| Учреждение (`institution`) | Управление профилем, аналитика, чат, вакансии, просмотр откликов и резюме, ответ на отзыв | Владелец учреждения (верифицированный, FR-34) | MVP |
| Модератор (`moderator`) | Очередь регистраций (approve/reject), модерация отзывов, разрешение споров (FR-35) | Сотрудник EduHub | MVP |
| **Администратор МОН** (`admin`) *(решение №4 — объединённая роль)* | Реестр и верификация учреждений, агрегированная аналитика по регионам (FR-24) **+** тарифы Pro/Enterprise, maintenance-режим платформы | МОН РТ либо делегированный сотрудник EduHub с этим доступом | MVP / Core |

\* Право оставить отзыв — не роль, а факт: `children.some(c => c.instId === inst.id)` (гейт уже реализован в `institutions/[id]/page.tsx`, SRS FR-15).

### 1.2 Профильные факты (не роли, отдельные сущности/поля User)

| Факт | Определяется | Даёт доступ к |
|---|---|---|
| Родитель | `Child[].length > 0` | Написание отзыва (при привязке к конкретному учреждению), вкладка «Дети» |
| Соискатель | Наличие/видимость `Applicant`-профиля | `/account` → «Профиль», публикация резюме, отклики на вакансии, обращения от учреждений |

### 1.3 Критично для бэкенда: сейчас RBAC нет вообще

Во фронте нет `middleware.ts` и ни одного route-level гейта — `/dashboard`, `/moderator`, `/admin` открыты по прямому URL без проверки роли. Единственная «защита» — NavBar показывает нужную кнопку по `role` из localStorage. Это ожидаемо для чистого прототипа, но **реальный RBAC-middleware на каждом API-эндпоинте — обязательная и первая по приоритету бэкенд-задача** (SRS §10 уже это требует).

---

## 2. Полный функциональный каталог

Формат: **FR** — описание → сущности → роли → экран(ы) фронта → статус.

### 2.1 Каталог и поиск

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-01 | Полнотекстовый поиск по названию/району/типу | Institution | Гость+ | `/search` | ✅ |
| FR-02 | Фасетные фильтры: район, тип(ы) (мульти), программа, ступень, возраст, цена, языки, развозка, питание, скидки, рейтинг | Institution | Гость+ | `/search`, URL=источник истины (shareable) | ⚠️ частично — тип/регион/район/цена/рейтинг/транспорт/питание/verified есть; **мульти-тип, программа-обучения-в-фильтре, ступень, скидки — поля есть не везде, см. §5.1** |
| FR-03 | Геопоиск «рядом со мной» с деградацией GPS→регион→страна | Institution.geo | Гость+ | `/` (секция «рядом с вами»), `/map` | ✅ честная деградация (`lib/geo.ts`, `haversine`) |
| FR-04 | Сравнение 2-3 учреждений + экспорт PDF | Institution | Гость+ | — | ❌ **исключено из MVP** (решение №2, было убрано во 2-й волне) — Core/вторая волна |
| FR-05 | Избранное и сохранённые поиски | `savedIds[]` | user | Везде (сердце на карточке), `/account` | ⚠️ избранное ✅; «сохранённые поиски» (persist фильтров как именованный запрос) — ❌ нет |

### 2.2 Профиль учреждения

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-06 | Карточка: контакты, лицензия, программы, языки, ориентиры расположения | Institution | Гость+ | `/institutions/[id]` (О нас) | ⚠️ контакты/адрес/описание ✅; `license_no`, `languages[]`, `location_landmarks` — ❌ полей нет во фронте |
| FR-07 | Галерея фото + видеотур | Institution.gallery | Гость+ | вкладка «Галерея» | ⚠️ фото ✅; видеотур — ❌ нет |
| FR-08 | Педагоги: ФИО/должность/образование/опыт/фото | Person | Гость+ | вкладка «Персонал», `/people/[id]` | ✅ |
| FR-09 | Кружки и достижения | Achievement | Гость+ | вкладка «Достижения» | ✅ |
| FR-10 | Финансовый калькулятор (полная стоимость + скидки) | Institution.price + discount* | Гость+ | — | ❌ **нет** — только статичная цена, калькулятора с учётом скидок нет |
| FR-11 | Инклюзивность (доступность, психолог, спецподдержка) | Institution.metrics | Гость+ | «Рейтинг по параметрам» | ⚠️ метрика-оценка есть, отдельного описательного блока «что именно доступно» нет |
| FR-12 | Развозка: тип/стоимость/районы | Institution.transportInfo | Гость+ | модалка «Транспорт» | ✅ |
| FR-13 | Питание: наличие/тип/стоимость (**без халяль**, решение №3) | Institution.foodMenu, meals* | Гость+ | вкладка «Питание» | ✅ соответствует пересмотренному FR-13 |
| FR-26 | Скидки/стипендии: наличие, тип, описание — разовое поле при регистрации | `discountAvailable/discountType/discountDetails` | Гость+ (просмотр), institution (заполнение) | — | ❌ **отсутствует во фронте** — есть в SRS Приложение A v2.1, не реализовано ни разу до этой сессии |
| FR-27 | Мульти-тип учреждений (комплексы, сад+школа и т.п.) | `Institution.types: CategoryKey[]` | institution (регистрация), гость (фильтр/категории) | — | ❌ **отсутствует во фронте** — сейчас `Institution.type` (единичное значение) + дублирующее `tk` |
| FR-28 | Ступень/программа/направление (школа: 1-4/5-9/10-11; вуз: бакалавр/магистр/докторант; центр: направление) | `Institution.programLevel: ProgramLevelKey[]` | Гость+ | — | ❌ **отсутствует во фронте** — есть только `curriculum` (гос/билингва/международная/STEM), это другое поле |

### 2.3 Рейтинги и отзывы

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-14 | Оценка по метрикам (1-5★ каждая) — **см. §5.2, факт 8, не 9** | `Institution.metrics[8]`, `Review.metrics[8]` | Гость+ (просмотр), user (оценка) | «Рейтинг по параметрам», форма отзыва | ✅ |
| FR-15 | Верификация автора через привязку родитель–ребёнок–учреждение | `ChildLink`/Child | user | Форма отзыва (гейт), «Дети» | ✅ (плюс 3-й статус `transferred`, сверх SRS) |
| FR-16 | Обязательная модерация ≤24-48ч перед публикацией | Review.status | moderator | — | ❌ **нет очереди модерации отзывов** — есть только очередь модерации регистраций учреждений (`/moderator`) |
| FR-17 | Ответ учреждения на отзыв | `Review.reply` | institution | вкладка «Отзывы» дашборда | ✅ |
| FR-18 | Антиспам, защита от накрутки (IP/устройство/частота, детекция аномалий) | — | Система | — | ❌ backend-only, неприменимо к прототипу |
| FR-29 | Порог минимума отзывов до показа агрегированного рейтинга | — | Система | — | ❌ Core-фаза по SRS, ожидаемо |
| FR-30 | Затухание веса отзывов по давности (recency decay) | — | Система | — | ❌ Core-фаза, формула — предмет отдельного дизайна (SRS сама это оговаривает) |
| FR-34 | Верификация владельца профиля учреждения при регистрации | `Institution.verified`, `Institution.status` | moderator | `/moderator` | ⚠️ **частично** — approve/reject реализован (коммит `3b8eff9`), но нет `verified_by`/`verified_at` (audit trail отсутствует, см. §5.4) |
| FR-35 | Формальный процесс разрешения споров по отзывам, SLA ≤72ч | Review.dispute_status | institution, moderator | — | ❌ отсутствует (Core-фаза) |

### 2.4 Коммуникации

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-19 | Чат родитель↔учреждение с историей сообщений (WebSocket) | Conversation, Message | user, institution | `/messages` | ⚠️ UI полностью есть, данные — статичный сид (`CONVS`/`MESSAGES` в отдельном `localStorage`), нет WebSocket/сервера |
| FR-20 | Уведомления: push, SMS, email | Notification | user, institution | Колокольчик в NavBar | ⚠️ только in-app (localStorage), push/SMS/email — backend-only, не реализовано |
| FR-21 | Реферальная система | — | user | — | ❌ отсутствует (Core-фаза по SRS) |
| FR-36 | Базовая доска вакансий: публикация (учреждение) / отклик (соискатель) | Vacancy, Application | institution, user-соискатель | `/vacancies`, `/dashboard` | ✅, включая идемпотентность отклика (`hasApplied`) |
| **FR-37** *(новое, решение №1)* | Публичные резюме соискателей + прямое обращение учреждения к соискателю (обратный поиск, лёгкая версия — без фильтров/поиска по базе, просто публичный список + возможность написать) | Applicant (`visibility=public`), EmployerResponse | institution, user-соискатель | `/applicants`, `/applicants/[id]` | ✅ построено (12-я волна), формально закреплено этим решением в MVP-скоуп |

### 2.5 Аналитика

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-22 | Дашборд учреждения: просмотры, источники трафика, клики | — (events, не смоделировано) | institution | `/dashboard` (Обзор, AreaChart) | ⚠️ визуализация есть (`recharts`), данные не реальные — нет трекинга событий |
| FR-23 | Экспорт отчётов CSV/PDF | — | institution | — | ❌ отсутствует |
| FR-24 | Агрегированная аналитика для МОН по регионам | — | admin (решение №4) | — | ❌ отсутствует, чисто backend + отдельный B2B-портал/API, вне скоупа текущего фронта |

### 2.6 Монетизация и платформа

| FR | Описание | Сущности | Роли | Экран | Статус фронта |
|---|---|---|---|---|---|
| FR-25 | Тарифы Free/Pro $30/Enterprise $100, платный тариф поднимает видимость в поиске, НЕ влияет на рейтинг | `Institution.plan` | institution, admin | `/admin` (только цены) | ❌ **гэп**: `Institution.plan` нет в типе данных вообще; продвижение по тарифу в сортировке `/search` не реализовано; `/admin` меняет только числа цен |
| *(вне FR-нумерации SRS)* | Модерация регистрации учреждений (pending→approved/rejected) | `Institution.status` | moderator/admin | `/moderator` | ✅ (коммит `3b8eff9`) |
| *(вне FR-нумерации SRS)* | Платформенные настройки (тарифы, maintenance-режим) | `PlatformSettings` | admin | `/admin` | ✅ |

---

## 3. Модель данных — сущности для Go/Postgres backend

Конвенции проекта (`.claude/rules/db.md`): UUID PK, `TIMESTAMPTZ`, `GEOGRAPHY(Point,4326)` + GIST-индекс для гео, двуязычные поля — `JSONB {ru, tg}` (симметрично `lib/i18n.tsx` на фронте).

### 3.1 User — отсутствует в прототипе вообще, фундаментальный гэп №1

Фронт хранит всю «учётку» одним blob'ом в `localStorage["eduhub_app_state_v2"]` — нет реальных аккаунтов, паролей, сессий. SRS §8.1 уже специфицирует JWT (access+refresh). **Это первая задача бэкенда** — всё остальное держится на ней.

```
User {
  id              UUID PK
  phone_or_email  String UNIQUE
  password_hash   String                                  // bcrypt/argon2
  role            Enum [user|institution|moderator|admin]  // guest = отсутствие сессии
  locale          Enum [ru|tg]
  region          Region NULL        // последний выбранный регион; GPS-координаты — только in-memory на клиенте, не персистить (PII-минимизация)
  institution_id  UUID NULL FK -> Institution.id  // если role=institution
  created_at      Timestamp
}
```

### 3.2 Institution (с учётом FR-26/27/28, ещё не реализованных во фронте)

```
Institution {
  id                 UUID PK
  name               JSONB {ru,tg}
  types              Enum[] [kindergarten|preschool|school|center|university]  NOT NULL, min 1  // FR-27
  region             Enum [dushanbe|sughd|khatlon|gbao|rrp]
  city, area, street JSONB {ru,tg} / String
  geo                GEOGRAPHY(Point,4326)  NOT NULL       // FR-03, GIST-индекс
  location_landmarks String                                // FR-06 — гэп, нет во фронте
  price              Money
  founded            Int
  students           Int
  age_range          String
  description        JSONB {ru,tg}
  license_no         String                                // FR-06 — гэп
  languages          String[]                               // FR-06 — гэп
  curriculum         Enum[] [state|bilingual|international|stem]
  program_level      Enum[] [primary|basic|secondary|bachelor|master|doctorate|dir_languages|dir_it|dir_sport|dir_creative|dir_stem|dir_other]  // FR-28 — гэп
  transport_type     Enum [own_bus|minibus|taxi|parent_coop|none]
  transport_cost     Money NULL
  transport_areas    String[] NULL
  meals_available    Enum [yes|no|bring_own]
  meals_type         Enum [hot|breakfast|buffet|none] NULL
  meals_cost         Money NULL
  -- meals_halal НЕТ (решение №3) --
  discount_available Boolean  DEFAULT false                 // FR-26 — гэп
  discount_type      Enum[] [needs_based|merit_or_competition|large_family|other] NULL
  discount_details   JSONB {ru,tg} NULL
  status             Enum [pending|approved|rejected]        // модерация регистрации
  verified           Boolean  DEFAULT false                  // FR-34
  verified_by        UUID NULL FK -> User.id                 // audit trail — гэп, нет во фронте
  verified_at        Timestamp NULL
  plan               Enum [free|pro|enterprise]  DEFAULT 'free'  // FR-25 — гэп, поля нет во фронте вообще
  score, review_count Float/Int                              // денормализованный агрегат, пересчитывается из RatingScore
  owner_user_id      UUID FK -> User.id
  created_at         Timestamp
}
```

### 3.3 Child (SRS §7, минимизация PII — обязательный API-контракт)

SRS: только `age_group`+`status`, **никогда** имя/фото/дата рождения ребёнка. Фронт (`ChildLink`) хранит `name` как client-only удобство для родителя — **этот `name` не должен уходить на бэкенд**.

```
Child {
  id             UUID PK
  user_id        UUID FK -> User.id
  institution_id UUID FK -> Institution.id
  age_group      Enum [kindergarten|preschool|primary|basic|secondary|university]
  status         Enum [current|alumnus|transferred]  // 3-е значение — фронтовое расширение, для FR-15 приравнивается к alumnus (связь была реальной)
  created_at     Timestamp
}
```
**API-контракт:** `POST /api/v1/children` принимает только `institution_id, age_group, status` — без `name`. Закрепить явно в API-документации, чтобы имя ребёнка не попало в DTO по инерции при реализации.

### 3.4 Review + RatingScore (метрики — 8, не 9, см. §5.2)

Метрики НЕ выделяются как JSONB-массив на Review — нормализованная таблица нужна для FR-29 (порог) и FR-30 (затухание по давности), которые требуют агрегации по времени/по метрике:

```
Review {
  id             UUID PK
  user_id        UUID FK -> User.id
  institution_id UUID FK -> Institution.id
  child_id       UUID FK -> Child.id                 // верификация FR-15
  text           JSONB {ru,tg}
  status         Enum [pending|published|rejected]    // FR-16
  reply          JSONB {ru,tg} NULL                    // FR-17
  dispute_status Enum [none|escalated|resolved_kept|resolved_removed] NULL  // FR-35
  created_at     Timestamp
}

RatingScore {
  id          UUID PK
  review_id   UUID FK -> Review.id
  metric_key  Enum [quality|conditions|safety|food|transport|price|parent_involvement|inclusivity]  // 8 метрик, см. §5.2
  value       SmallInt CHECK (value BETWEEN 1 AND 5)
}
```
`Institution.score` — вычисляемый агрегат (по метрике `quality`+7 остальных → общий), не независимо оцениваемая 9-я метрика (SRS/бизнес-план формулируют это неточно, см. §5.2).

### 3.5 Teacher/Person, Achievement, Alumnus, NewsArticle, Gallery — 1:N к Institution

Прямой перенос из фронтовой схемы без изменения логики:
- `Person { id, institution_id, name, role_label, role_type[director|teacher|coach|psychologist|admin], subject?, photo, experience, bio, education[], email?, phone? }` + `Achievement { id, owner_type[institution|person], owner_id, title, year, category[gold|silver|bronze|special], description, links[] }` — полиморфный владелец.
- `Alumnus { id, institution_id, name, photo, grad_year, current_occupation }` — используется только для типов `school`/`university` (SRS не содержит эту сущность вообще, чисто фронтовое расширение с прошлых волн — низкий приоритет, не привязано ни к одному FR, можно реализовать в Core-фазе или позже).
- `NewsArticle { id, institution_id, title, category, cover_url, video_url?, content, tags[], status[published|draft], views, created_at }` — тоже не имеет собственного FR в SRS (косвенно упомянуто в бизнес-плане §05 «постинг новостей»), стоит формально добавить как FR-38 при следующем обновлении SRS.
- `gallery: {url, label}[]` — сейчас инлайн-массив на Institution, при переносе на бэкенд разумно вынести в отдельную таблицу `GalleryItem{id, institution_id, url, label, sort_order}` для нормального CRUD и лимитов на количество.

### 3.6 Vacancy / Applicant / Application / EmployerResponse (FR-36 / FR-37)

```
Vacancy { id, institution_id, title, description, requirements[], salary_from?, salary_to?, employment, status[published|draft], created_at }

Applicant {                                    // 1:1 к User, факт профиля "соискатель"
  id, user_id UNIQUE,
  name, photo, position, bio JSONB{ru,tg},
  education[], experience[], skills[] JSONB{ru,tg}[],
  cv_file_url  String NULL,                     // реальный S3-аплоад — гэп, сейчас только cvFileName без файла
  visibility   Enum [draft|on_response|public],
  hide_contacts Boolean,
}

Application {                                  // отклик соискатель → вакансия, FR-36
  id, applicant_id, vacancy_id, status[sent|viewed|closed], created_at,
  UNIQUE(applicant_id, vacancy_id)              // идемпотентность — уже так в app-state.tsx hasApplied()
}

EmployerResponse {                              // FR-37 (решение №1), обращение institution → applicant
  id, institution_id, applicant_id, message JSONB{ru,tg}, created_at,
  UNIQUE(institution_id, applicant_id)          // идемпотентность — уже так в hasResponded()
}
```
**Рефакторинг при переносе:** фронтовая `VacancyCandidate` (кто откликнулся на конкретную вакансию, вид со стороны учреждения) — это производное представление `Application JOIN Vacancy WHERE vacancy.institution_id = X`, отдельной таблицей на бэке быть не должно.

### 3.7 Message/Conversation, Notification, ModerationLog, PlatformSettings

```
Conversation { id, user_id, institution_id, created_at, UNIQUE(user_id, institution_id) }
Message { id, conversation_id, sender_type[user|institution], text, sent_at }   // WS /ws/chat/{institutionId}, FR-19

Notification { id, user_id, type[message|review_reply|news|achievement], text, link?, read, created_at }  // FR-20

ModerationLog {                                 // НОВОЕ — architecture.md требует audit trail модерации, сейчас нигде не реализовано
  id, moderator_id UUID FK->User.id,
  target_type Enum [institution_registration|review_dispute],
  target_id   UUID,
  action      Enum [approved|rejected|verified],
  reason?     String,
  created_at  Timestamp
}

PlatformSettings { id = 1 (singleton), tier_price_pro Money, tier_price_enterprise Money, maintenance_mode Boolean }  // admin, FR-25 частично
```

### 3.8 Client-only демо-хаки — НЕ проектировать 1:1 на бэке

- `Applicant.cv_file_url` — сейчас только имя файла (`cvFileName`), нет S3/бинарника — нужен реальный upload-flow (антивирус-скан, лимит размера, per `.claude/rules/security.md`).
- Фото профиля соискателя — сейчас `FileReader.readAsDataURL()` → base64 в состоянии, никуда не уходит — нужен реальный upload в S3+CDN.
- `registerInstitution()` — сейчас мгновенный локальный insert (`id = max+1`) — на бэке обязателен реальный `INSERT` + очередь модерации (уже частично смоделирована через `status`).
- `DASH_INST_ID = 1` (хардкод в `/dashboard`) — demo-fallback вместо честного empty-state; на бэке — привязка `user.institution_id` через auth-сессию, без констант.
- `CONVS`/`MESSAGES` — статичный сид, никакого WebSocket; на бэке — реальный `Conversation`/`Message` + WS-сервер (FR-19).

---

## 4. PII и минимизация данных (для раздела безопасности бэкенда)

- **`Child.name`/`age`** — фронт хранит имя ребёнка только локально (client-only удобство родителя), **никогда не должно уходить на сервер** (см. API-контракт §3.3).
- **`Applicant.email/phone`** + `hide_contacts` — уже встроенное архитектурное решение по минимизации (скрыть контакты соискателя до ответа учреждения в чате) — перенести логику как есть.
- **`Applicant.photo`**, реальные фото/CV при аплоаде — PII, требует контроля доступа на уровне S3 (не публичные bucket'ы без подписанных URL).
- **`Institution.address`/`street`/`geo`** — точный адрес+координаты учреждения, не человека, но чувствительно в контексте «где находится ребёнок» — доступ read-only публичный (это нормально, адрес учреждения не является PII по Закону РТ №1537, в отличие от данных о самом ребёнке).
- Соответствие Закону РТ №1537 — согласие на обработку ПД при регистрации, право на удаление аккаунта и связанных данных (включая `Child`, `Applicant`, `Review`) — отдельный API-эндпоинт `DELETE /api/v1/users/me` с каскадным удалением/анонимизацией.

---

## 5. Расхождения SRS/бизнес-план ↔ факт кода

### 5.1 Закрыто решениями №1-4 (см. §0)
FR-37 в MVP, FR-04 исключён из MVP, halal убран из FR-13, роль admin слита с «Администратор МОН».

### 5.2 «9 метрик» → «8 независимых + 1 вычисляемая» — расхождение с документами, требует правки текста

В таблице §06 бизнес-плана и FR-14 SRS фигурируют «9 независимых метрик», где 9-я — буквально «Общий рейтинг = среднее остальных 8» (не независимо оцениваемый параметр, а производная величина). Во фронте это уже исправлено сайтвайд (`web/lib/data.ts` — ровно 8 записей в `metrics[]` на каждой институции, проверено построчно). Корректная формулировка для будущего SRS v2.2/бизнес-плана v3.4: **«8 независимых метрик + 1 вычисляемый общий рейтинг»**. Метрики (точные лейблы из кода): Качество обучения, Условия и помещения, Безопасность, Питание, Развозка, Цена/качество, Участие родителей, Инклюзивность.

### 5.3 FR-25 / `Institution.plan` — гэп без конфликта, не требует решения пользователя сейчас

SRS предполагает поле `plan` на Institution, влияющее на видимость в поиске. Поля нет ни во фронт-типе, ни в сортировке `/search`. Реальный функционал монетизации (кроме `/admin`-настройки цен тарифов) отсутствует полностью. Это чисто объём работы для backend + frontend доработки, не архитектурная развилка — включено в фазовый план §6 как есть.

### 5.4 Отсутствующий `ModerationLog` — нарушение `.claude/rules/architecture.md`

Правило проекта требует audit trail действий модерации. Ни во фронте, ни в SRS §7 нет сущности, фиксирующей «кто/когда/что решил» при approve/reject учреждения или разрешении спора по отзыву. Добавлено в модель данных как новая сущность (§3.7) — требует явного FR-номера при следующем обновлении SRS (предлагается FR-39).

### 5.5 `NewsArticle`/`Alumnus`/`GalleryItem` — реализованы во фронте, но не имеют закреплённого FR

Публикация новостей упомянута в бизнес-плане §05 без номера FR. Alumni-раздел и структурированная галерея вообще не упомянуты ни в одном документе (чисто фронтовые расширения прошлых волн). Не блокирует бэкенд-план (перенос механический), но при следующей ревизии SRS стоит закрепить FR-38 (новости) и явно решить статус alumni (оставить как есть без FR, или формализовать).

---

## 6. Фазовый план (согласован с бизнес-планом §12, синхронизирован с решениями §0)

| Фаза | Период (бизнес-план) | Backend-скоуп |
|---|---|---|
| **MVP** | Мес. 1-3 ($15K) | User/Auth (JWT access+refresh), Institution CRUD + модерация регистрации + `verified` вручную (FR-34 — ручная проверка лицензии модератором), Child + Review без порога/decay, RatingScore (8 метрик), Vacancy/Application, **EmployerResponse (FR-37)**, FR-26/27/28 (скидки/мульти-тип/ступень), ModerationLog, статический каталог+поиск (Postgres full-text, без Elasticsearch), геопоиск PostGIS, WebSocket-чат (реальный, не сид — FR-19 в MVP по SRS §4) |
| **Core** | Мес. 4-6 ($18-20K) | Notification (push/SMS/email), FR-29/30 (порог+decay — формула нуждается в отдельном дизайне перед стартом фазы), FR-35 (споры+SLA), FR-21 (реферальная), FR-22/23 (реальная аналитика+экспорт), **FR-04 (сравнение, перенесено сюда решением №2)**, Elasticsearch (если рост данных оправдывает) |
| **Вторая волна** | Мес. 7-9 ($25K) | AI-подбор (weighted scoring по анкете), полноценная CV-база с фильтрами/поиском по резюме (сверх FR-37 lite-версии), нативные приложения, платежи (Корти Милли/Alif Pay), онлайн-эквайринг премиум-тарифов |
| **Scale** | Мес. 10-12 ($15K) | API для МОН (FR-24, B2B/аналитический портал), Kubernetes (если нагрузка оправдывает, `.claude/rules/devops.md`), расширение регионов/языков (узб., англ. по NFR) |

---

## 7. Остающиеся открытые вопросы (не блокируют начало бэкенда, решать по мере приближения к фазе)

1. **Формула FR-29/30** (порог отзывов, коэффициент затухания по давности) — SRS сознательно оставляет как «предмет Core-дизайна». Нужна отдельная сессия с конкретными цифрами перед стартом Core-фазы, не раньше.
2. **Обновлять ли `.docx` (SRS v2.2 / бизнес-план v3.4)** — сейчас решено не трогать, этот файл — временный журнал синхронизации (см. §0, последний пункт). Пересмотреть, когда/если в проект придёт постоянная бэкенд-команда, которой нужен единый читаемый источник правды без двух параллельных документов.
3. **FR-38/FR-39** (новости, ModerationLog) — предложены новые номера FR для следующей ревизии SRS, не присвоены официально нигде, кроме этого документа.
4. **Alumni-раздел** — оставить неформализованным фронтовым расширением или закрепить как FR при следующей ревизии SRS — открыто, низкий приоритет.
