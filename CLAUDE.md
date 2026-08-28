# EduHub — Project Contract

## Что это
EduHub Tajikistan — национальная платформа поиска, сравнения и оценки образовательных учреждений (сады, дошкольные центры, школы, центры допобразования, вузы). Трёхсторонний рынок: родители/студенты, учреждения, соискатели-педагоги (соискатели — вторая волна). MVP — сразу весь Таджикистан (не только Душанбе), архитектурно регион-агностично с самого начала (5 регионов уже поддержаны фильтром/сидами: Душанбе/Сугд/Хатлон/ГБАО/РРП).

Источники истины:
- `docs/EduHub_Technical_Specification.docx` — SRS: изначальная архитектура, стек, модель данных, NFR
- `docs/EduHub_Business_Plan_Pro.docx` — бизнес-модель, рынок, роадмап, KPI
- `docs/EduHub_Functional_Requirements.md` — актуальный список функциональных требований (FR-01…FR-41), пополняется при добавлении новой фичи
- `docs/EduHub_Backend_Architecture.md` — архитектурные решения бэкенда (точка входа, ссылается на схему БД, план вех, тарифы)
- `web/` — фактический Next.js-фронт как визуальный референс (`EduHub_Prototype.html` в репозитории больше не существует)

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
- **Pixel-parity**: при портировании UI-паттернов из существующего `web/` — без поведенческих изменений, если не оговорено явно
- **Scope gate**: перед реализацией новой фичи сверяться с FR-списком

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
