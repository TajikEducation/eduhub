# EduHub — Project Contract

## Что это
EduHub Tajikistan — национальная платформа поиска, сравнения и оценки образовательных учреждений (сады, дошкольные центры, школы, центры допобразования, вузы). Трёхсторонний рынок: родители/студенты, учреждения, соискатели-педагоги (соискатели — вторая волна). MVP — сразу весь Таджикистан (не только Душанбе), архитектурно регион-агностично с самого начала (5 регионов уже поддержаны фильтром/сидами: Душанбе/Сугд/Хатлон/ГБАО/РРП).

Источники истины:
- `docs/EduHub_Technical_Specification.docx` — SRS: архитектура, стек, FR-01…FR-24, модель данных, NFR
- `docs/EduHub_Business_Plan_Pro.docx` — бизнес-модель, рынок, роадмап, KPI
- `EduHub_Prototype.html` — визуальный референс (pixel-parity источник для UI)

## Stack (по SRS, раздел 5)
- Backend: Go (сервис-ориентированная архитектура, REST/JSON + WebSocket для чата)
- Frontend: React + Next.js (SSR ради SEO)
- Стили: Tailwind CSS
- БД: PostgreSQL + PostGIS (гео-поиск «рядом со мной»)
- Кэш/очереди: Redis
- Поиск: Elasticsearch (подключается по мере роста; для MVP может хватить поиска на стороне PostgreSQL)
- Медиа: S3 + CDN
- Инфра: Docker (Kubernetes — после MVP, по мере нагрузки)

## Скоуп-правило MVP (SRS §2 — обязательно к соблюдению)
Если для поддержания данных учреждению нужно делать что-то **регулярно** (вести GPS, обновлять меню, проводить оплату) — это операционная надстройка → **вторая волна**, не реализовывать без явного запроса.
Если учреждение заполняет поле **один раз** при регистрации — это информационное поле → **входит в MVP**.

Следствия:
- Развозка и питание — поля `Institution` (`transport_type/cost/areas`, `meals_available/type/cost/halal`), НЕ отдельные сервисы, НЕ GPS в реальном времени, НЕ конструктор меню.
- Вне MVP пока не запрошено явно: GPS-трекинг транспорта, онлайн-оплата, доска вакансий, нативные мобильные приложения (см. SRS §11).

## Роли (RBAC, SRS §3)
Гость → Родитель/студент → Учреждение → Модератор → Администратор платформы (MVP/Core). Соискатель — вторая волна.

## Ключевые сущности (SRS §7)
`User`, `Institution` (вкл. поля развозки/питания), `Rating` (8 метрик — девятая величина в ранних формулировках была вычисляемым средним, не отдельным параметром, см. FR-14/29/30), `Review`, `Teacher`, `Message`. Точная схема `Institution` — см. SRS §7.2, не выдумывать поля мимо неё.

## Требования трассируются к FR-XX
Функциональные требования MVP пронумерованы FR-01…FR-24 (SRS §6). При реализации фичи — ссылаться на соответствующий FR; если фича не покрыта ни одним FR и не относится к явному запросу пользователя — уточнить, не реализовывать по умолчанию.

При добавлении нового FR — обязательно дописать его в `docs/EduHub_Functional_Requirements.md` (формат: Цель → Что нужно реализовать → Пример → Реализация в коде → Резюме простыми словами; без ссылок на внешние документы/файлы/секции).

## NFR, которые нельзя игнорировать (SRS §9–10)
- API p95 ≤ 300 мс; Uptime ≥ 99.5%
- Mobile-first (80%+ трафика с телефонов), PWA с offline-кэшем просмотренных учреждений
- Локализация: тадж. + рус. с самого начала (архитектурно, не «добавим потом»)
- Минимизация данных о детях; контактные данные — только по необходимости
- Пароли — bcrypt/argon2; TLS everywhere; защита от OWASP Top 10; audit log действий модерации

## NEVER
- Commit/push без явного запроса пользователя
- Добавлять `Co-authored-by` без явного запроса
- Коммитить секреты/API-ключи (`.env`, `.claude/settings.local.json`)
- Реализовывать функции второй волны (GPS-трекинг, онлайн-оплата, доска вакансий, нативные приложения) без явного запроса — это осознанное решение по бюджету/срокам, не забывчивость
- Хранить/логировать лишние данные о детях сверх необходимого
- Ломать публичные API/контракты без явного запроса
- Менять `.env`, lockfiles, CI secrets без подтверждения

## Mandatory Behavior
- **Язык**: диалоги и пояснения — на русском (технические термины — английский допустим); коммиты — conventional commits на английском
- **Explain before change**: объяснить что/почему до правки кода
- **Checkpoint**: при "продолжаем" / "resume" читать `CLAUDE.local.md`
- **Transparency**: конец каждого ответа — список использованных категорий инструментов
- **Pixel-parity**: при портировании из `EduHub_Prototype.html` — без поведенческих изменений, если не оговорено явно
- **Scope gate**: перед реализацией новой фичи сверяться со скоуп-правилом выше и FR-списком

## Compact Instructions
When compressing, preserve in priority order:
1. Architecture decisions (NEVER summarize away)
2. Modified files and key changes
3. Verification status — which commands passed/failed
4. Open TODOs and rollback notes
5. Tool outputs — delete freely, keep only pass/fail summary

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
