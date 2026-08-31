# Graph Report - eduhub  (2026-08-30)

## Corpus Check
- 169 files · ~234,709 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1486 nodes · 3184 edges · 159 communities (103 shown, 56 thin omitted)
- Extraction: 89% EXTRACTED · 11% INFERRED · 0% AMBIGUOUS · INFERRED: 354 edges (avg confidence: 0.8)
- Token cost: 703,709 input · 0 output

## Community Hubs (Navigation)
- Catalog Repo & Transport Tests
- Business Plan Sections
- HTTP Error Taxonomy (apperr)
- Devseed Demo Data
- Web Search/Vacancies Pages
- Dashboard CRUD Actions
- Dashboard Components & Forms
- Design Taste Frontend Skills
- Web NPM Dependencies
- Applicant Profile Pages
- Register-Institution & Dashboard Layout
- About Page Content
- Cached Catalog Service
- Admin Dashboard Page
- Web TS Config
- Backend main() & Geo Repo Tests
- API Deps & Pinger Health Check
- apperr Constructors
- Health/Readyz Handler
- Clock Abstraction
- Dashboard Page Actions
- Company Page & Contact Form
- Access Log Middleware Tests
- Backend Architecture Rationale
- Catalog Filter Validation Tests
- CORS Middleware Tests
- Catalog Handler Tests
- Pick UI Library Skill
- Access Log Middleware
- Router & Logger Setup
- Apple Design Skill
- Institution Satellite Tables
- API Router Wiring
- HTTP Router & Chain
- Account Page (Web)
- Animation Recipes
- Catalog Query Parsing
- JSON Decode Helpers
- Geo Detection Utilities
- Animation Skills Bundle
- Milestone 1 Catalog Tasks
- Institution Get Handler & ETag
- Clock Package Tests
- Healthz/Readyz Handler Tests
- DB Schema Ownership & Employment
- Auth Schema Tables
- Backend Docker Compose Services
- Backend Config Loader
- i18n & Vacancy Filters
- Animation Audit Categories
- Devseed Main & Integration Test
- Config Loader Tests
- CLAUDE.md Rules & ADR-0001
- Design Engineering Philosophy Skills
- Rate Limit Middleware
- Login/Register OAuth Flow
- CORS Middleware
- Panic Recovery Middleware
- Site Layout & Footer
- Geo Locate Utility
- SMM Rating Rules
- Login Page (Web)
- Register Page (Web)
- Redis Cache Perf Baseline
- Access Log statusWriter
- SMM Site Audit Findings
- Root Layout & Fonts
- Sonner Toast Skill
- Verify Page (Web)
- Web README Boilerplate
- UI Design Recipes Misc
- Transactional Outbox Pattern
- Chat Window Helper
- Brandkit/Image-to-Code Skills
- Full Output Enforcement Skill
- Gallery & Media Upload FR
- Rating Sync
- Notifications FR
- Web Agent Rules Docs
- ESLint Config
- Next.js Config
- PostCSS Config
- Push Subscriptions FR
- Institution Dashboard Export FR
- Investor Numbers Open Question
- Auth Channels Decision
- Chat Generalization Decision
- Milestone 0 Checkpoint
- Frontend-Backend Contract Mismatches
- Milestone Critical Path
- Agent Delegation Lesson
- Search Boost Rework Decision
- Employer Reviews Decision
- Milestone 0 Tech Decisions
- Moderation Claim-Pattern Decision
- Modular Monolith Decision
- MON Role Removed Decision
- Observability Deferred Decision
- Chat Discovery Open Question
- Multipart Upload Decision
- Backend Go Module
- FR-16 Review Moderation
- FR-20 Notifications (RU)
- FR-21 Referral System
- Boilerplate file.svg Icon
- Boilerplate globe.svg Icon
- Guide Screenshot: Vacancy List
- Guide Screenshot: Vacancy Detail
- Guide Screenshot: Applicant Profile
- Guide Screenshot: Institution Overview
- Guide Screenshot: Institution Overview (2)
- Guide Screenshot: Parent Search
- Guide Screenshot: Institution Metrics
- Guide Screenshot: Parent Reviews Tab
- Guide Screenshot: Saved/Favorites
- Guide Screenshot: University Search
- Guide Screenshot: Alumni Tab
- Guide Screenshot: Student Reviews Tab
- Hero Image: Globe & Magnifying Glass
- Boilerplate next.svg Logo
- Boilerplate vercel.svg Logo
- Boilerplate window.svg Icon
- Web README (create-next-app)

## God Nodes (most connected - your core abstractions)
1. `useT()` - 68 edges
2. `useAppState()` - 45 edges
3. `C` - 42 edges
4. `FH` - 35 edges
5. `New()` - 33 edges
6. `New()` - 31 edges
7. `Institution` - 29 edges
8. `FB` - 26 edges
9. `SRS: EduHub Technical Specification (v2.3)` - 25 edges
10. `DashboardPage()` - 24 edges

## Surprising Connections (you probably didn't know these)
- `Animate Skill` --semantically_similar_to--> `Apple Design Skill`  [INFERRED] [semantically similar]
  agent/skills/animate/SKILL.md → agent/skills/apple-design/SKILL.md
- `Animation Vocabulary Glossary` --semantically_similar_to--> `Design Engineering (Emil Kowalski Philosophy)`  [INFERRED] [semantically similar]
  agent/skills/animation-vocabulary/SKILL.md → agent/skills/emil-design-eng/SKILL.md
- `ADR-0001: Separate Vacancy Detail Page` --cites--> `EduHub Project Contract (CLAUDE.md)`  [EXTRACTED]
  docs/adr/0001-vacancy-detail-page.md → CLAUDE.md
- `EduHub Project Contract (CLAUDE.md)` --cites--> `EduHub Backend Architecture Document`  [EXTRACTED]
  CLAUDE.md → docs/EduHub_Backend_Architecture.md
- `EduHub Project Contract (CLAUDE.md)` --cites--> `EduHub Functional Requirements`  [EXTRACTED]
  CLAUDE.md → docs/EduHub_Functional_Requirements.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Rating Integrity & Anti-Fraud Mechanisms** — srs_fr_18_31_antifraud, srs_fr_29_review_threshold, srs_fr_30_recency_decay, srs_fr_34_owner_verification, srs_fr_15_review_verification [INFERRED 0.85]
- **MoN RT Removal From Product (role, integration, personas, analytics)** — srs_mon_role_removed, bp_mon_removed, srs_roles_rbac [EXTRACTED 1.00]
- **Institution One-Time Informational Fields (Scope Rule)** — srs_transport_as_info_field, srs_meals_as_info_field, srs_discounts_as_info_field, srs_entity_institution [INFERRED 0.90]
- **Catalog Institution Card Satellite Tables (Composed via LATERAL json_agg)** — db_schema_catalog_institutions, db_schema_catalog_achievements, db_schema_catalog_institution_alumni, db_schema_catalog_institution_transport_routes [EXTRACTED 0.90]
- **Moderation Flow (Audit Log, Queue, Claim Pattern, State Machine)** — backend_dev_plan_moderation_design, db_schema_moderation_audit_log, db_schema_moderation_queue_items, db_schema_catalog_institution_owner_verifications, functional_requirements_fr16_17_35_moderation [INFERRED 0.85]
- **SMM Content Plan Restrictions from Broken Site Features** — smm_content_plan_visit_request_form_broken, smm_content_plan_hero_chips_broken, smm_content_plan_fake_amenities_histogram, smm_content_plan_about_page_ban [EXTRACTED 0.90]

## Communities (159 total, 56 thin omitted)

### Community 0 - "Catalog Repo & Transport Tests"
Cohesion: 0.06
Nodes (72): fakeCatalogRepo, Bilingual, Filter, ListResult, Achievement, Alumnus, GalleryItem, Institution (+64 more)

### Community 1 - "Business Plan Sections"
Cohesion: 0.05
Nodes (78): BP §14 Видение на 5 лет, BP Приложение A — журнал изменений v3.3, BP §09 Конкурентный анализ (FindSchool.kz, Eduland.uz), BP §01 Резюме проекта, BP §10 Финансовая модель (доход/расходы/прибыль год 1), BP §04 Типы образовательных учреждений, BP §11 Ключевые показатели (КПЭ), BP §09 Рынок TAM/SAM/SOM (пересчитано v3.3) (+70 more)

### Community 2 - "HTTP Error Taxonomy (apperr)"
Cohesion: 0.08
Nodes (40): Conflict(), Error, Forbidden(), Internal(), Invalid(), NotFound(), RateLimited(), sentinel (+32 more)

### Community 3 - "Devseed Demo Data"
Cohesion: 0.11
Nodes (42): newsForInst(), staffForInst(), biJSON(), biJSONPtr(), biSliceJSON(), ensureEntity(), pgx.Tx, insertAchievement() (+34 more)

### Community 4 - "Web Search/Vacancies Pages"
Cohesion: 0.09
Nodes (23): AboutPage(), MapPage(), ModeratorPage(), CHOICES, OnboardingPage(), CATEGORY_KEYS, HomePage(), CATEGORY_KEYS (+15 more)

### Community 5 - "Dashboard CRUD Actions"
Cohesion: 0.07
Nodes (25): bi(), newAchievement(), newAlumnus(), newArticle(), newGalleryItem(), newStaff(), newVacancy(), priceLabel() (+17 more)

### Community 6 - "Dashboard Components & Forms"
Cohesion: 0.06
Nodes (31): inputStyle, PHOTO_KEYS, SettingsTab, Tab, TAB_META, WEEK_TREND, CAT_COLOR, CAT_ICON (+23 more)

### Community 7 - "Design Taste Frontend Skills"
Cohesion: 0.06
Nodes (39): Design Taste Frontend Skill (v2), Design Taste Frontend v1 Skill, Imagegen Frontend Mobile Skill, Imagegen Frontend Web Skill, Industrial Brutalist UI Skill, Stitch DESIGN.md (Taste Standard), Stitch Design Taste Skill, Brief Inference Process (+31 more)

### Community 8 - "Web NPM Dependencies"
Cohesion: 0.05
Nodes (36): eslint, eslint-config-next, lucide-react, next, react, react-dom, recharts, tailwindcss (+28 more)

### Community 9 - "Applicant Profile Pages"
Cohesion: 0.10
Nodes (23): ACH_TIER, ApplicantProfilePage(), ApplicantsPage(), TiersCard(), InstitutionProfileInner(), loadMessages(), MessagesInner(), VacancyDetailPage() (+15 more)

### Community 10 - "Register-Institution & Dashboard Layout"
Cohesion: 0.09
Nodes (27): CATEGORY_KEYS, EMPTY, inputStyle, RegisterInstitutionPage(), AppStateContext, AppStateProvider(), AppStateValue, ChildLink (+19 more)

### Community 11 - "About Page Content"
Cohesion: 0.11
Nodes (18): MARKET, ROADMAP, SIDES, TRUST_PILLARS, VISION, TIERS, GuidePage(), MOTIFS (+10 more)

### Community 12 - "Cached Catalog Service"
Cohesion: 0.15
Nodes (20): NewCachedService(), discardLogger(), newFakeCache(), strPtr(), TestCache_50ParallelIdenticalRequests_RepoCalledOnce(), TestCache_CacheGetError_DegradesToRepo(), TestCache_DifferentFilters_DifferentKeys(), TestCache_SecondIdenticalRequest_DoesNotReachRepo() (+12 more)

### Community 13 - "Admin Dashboard Page"
Cohesion: 0.09
Nodes (21): AdminPage(), inputStyle, Tab, ACH_CATEGORIES, ACH_TIER, SOCIAL_META, Tab, NewsArticlePage() (+13 more)

### Community 14 - "Web TS Config"
Cohesion: 0.07
Nodes (28): dom, dom.iterable, esnext, **/*.mts, .next/dev/types/**/*.ts, next-env.d.ts, .next/types/**/*.ts, node_modules (+20 more)

### Community 15 - "Backend main() & Geo Repo Tests"
Cohesion: 0.17
Nodes (19): main(), assertNameOrder(), TestListGeo(), pgx.Tx, insertBareInstitution(), insertFullInstitution(), TestGetByID(), equalSets() (+11 more)

### Community 16 - "API Deps & Pinger Health Check"
Cohesion: 0.16
Nodes (16): Deps, fakePinger, Pinger, run(), TestRun_ClosesPoolAfterShutdown(), TestRun_GracefulShutdownWaitsForInFlightRequest(), TestRun_ServesHealthzOnEphemeralPort(), Open() (+8 more)

### Community 17 - "apperr Constructors"
Cohesion: 0.18
Nodes (21): Conflict(), Forbidden(), Internal(), RateLimited(), TestCategories_MatchSentinels(), TestInternal_UnwrapsCause(), TestMessage_FallsBackToCategoryWhenEmpty(), Unauthorized() (+13 more)

### Community 18 - "Health/Readyz Handler"
Cohesion: 0.15
Nodes (19): New(), Healthz(), pingWithTimeout(), Readyz(), TestHealth_Healthz_ReturnsOKWithoutDependencies(), TestHealth_Readyz_FailingPingerReturns503WithName(), TestHealth_Readyz_MultipleDependenciesOnlyFailingListed(), TestHealth_Readyz_RespectsTimeoutEvenWhenPingerIgnoresContext() (+11 more)

### Community 19 - "Clock Abstraction"
Cohesion: 0.15
Nodes (15): Clock, Fake, NewFake(), realClock, TestFake_AdvanceMovesNow(), TestFake_NowReturnsStart(), TestNew_ReturnsRealTime(), sync.Mutex (+7 more)

### Community 20 - "Dashboard Page Actions"
Cohesion: 0.12
Nodes (9): bi(), DashboardPage(), newAchievement(), newAlumnus(), newArticle(), newGalleryItem(), newStaff(), newVacancy() (+1 more)

### Community 21 - "Company Page & Contact Form"
Cohesion: 0.12
Nodes (17): AUDIENCE_LINKS, CompanyPage(), ContactForm(), PRINCIPLES, PROBLEMS, RATING_CATEGORIES, TOPICS, TRUST_PILLARS (+9 more)

### Community 22 - "Access Log Middleware Tests"
Cohesion: 0.24
Nodes (15): AccessLog(), TestAccessLog_DefaultStatusIsOK(), TestAccessLog_DoesNotLeakQueryString(), TestAccessLog_RequestIDMatchesContext(), TestAccessLog_WritesExpectedFields(), Recover(), TestRecover_CatchesPanicAndReturns500(), TestRecover_NoPanicPassesThrough() (+7 more)

### Community 23 - "Backend Architecture Rationale"
Cohesion: 0.13
Nodes (20): apperr Error Taxonomy, AuthN/AuthZ Design (argon2id, JWT reuse-detection, RBAC), Clean Architecture Layers (transport/usecase/domain/repo), EduHub Backend Architecture Document, Modular Monolith with Internal SOA Boundaries, Observability Deferred Past MVP, EduHub Backend Development Plan, Milestone 0 — Platform Skeleton (Tasks 1-16) (+12 more)

### Community 24 - "Catalog Filter Validation Tests"
Cohesion: 0.18
Nodes (16): T, p(), TestFilter_Normalize_CapsLimitAt50(), TestFilter_Normalize_DefaultLimit(), TestFilter_Validate_EmptyFilterIsValid(), TestFilter_Validate_LatWithoutLng(), TestFilter_Validate_MinPriceGreaterThanMaxPrice(), TestFilter_Validate_RadiusWithoutCoords() (+8 more)

### Community 25 - "CORS Middleware Tests"
Cohesion: 0.27
Nodes (10): CORS(), TestCORS_PreflightAllowedOriginReturns204WithHeaders(), TestCORS_PreflightDisallowedOriginNoAllowHeader(), TestCORS_RegularRequestAllowedOriginSetsHeaderAndCallsNext(), TestCORS_RegularRequestDisallowedOriginNoHeaderButCallsNext(), net/http.Header, statusProbe, writeRateLimitedError() (+2 more)

### Community 26 - "Catalog Handler Tests"
Cohesion: 0.18
Nodes (13): TestGetHandler(), floatPtr(), strPtr(), TestListHandler(), testLogger(), NotFound(), TestError_MessageNotEmpty(), TestInvalid_FieldsThroughAs() (+5 more)

### Community 27 - "Pick UI Library Skill"
Cohesion: 0.12
Nodes (16): Pick UI Library Skill, clsx, cmdk, Cobe, cva, dnd kit, input-otp, Leva (+8 more)

### Community 28 - "Access Log Middleware"
Cohesion: 0.23
Nodes (13): AccessLog(), TestAccessLog_DefaultStatusIsOK(), TestAccessLog_DoesNotLeakQueryString(), TestAccessLog_RequestIDMatchesContext(), TestAccessLog_WritesExpectedFields(), newRequestID(), RequestID(), TestRequestID_EmptyOnBareContext() (+5 more)

### Community 29 - "Router & Logger Setup"
Cohesion: 0.18
Nodes (13): NewRouter(), TestRouter_MatchedRouteServesNormally(), TestRouter_UnknownPathReturns404JSON(), TestRouter_WrongMethodReturns405JSON(), T, New(), parseLevel(), PtrOrNil() (+5 more)

### Community 30 - "Apple Design Skill"
Cohesion: 0.14
Nodes (15): Apple Design Skill, Animate Spring Config, Direct Manipulation (1:1 Tracking), Eight Design Principles, Interruptibility Principle, Materials & Depth (Translucency), Momentum Projection, Reduced Motion & Accessibility (+7 more)

### Community 31 - "Institution Satellite Tables"
Cohesion: 0.19
Nodes (14): Chosen Variant D: Satellite Tables (Mirrors institution_gallery), Task 17 — Migration 00002 catalog.institutions, Task 19 — Migration 00004 Satellite Tables, catalog.achievements, catalog.institution_alumni, catalog.institution_owner_verifications, catalog.institution_transport_routes, catalog.institutions (+6 more)

### Community 32 - "API Router Wiring"
Cohesion: 0.16
Nodes (9): catalogService, newHandler(), doGet(), TestSmoke_CatalogRoutesThroughRealServer(), Chain(), TestRouter_ChainCallsInDeclaredOrder(), net/http.Response, Middleware (+1 more)

### Community 33 - "HTTP Router & Chain"
Cohesion: 0.22
Nodes (9): Chain(), net/http.ServeMux, Router, main(), NewRouter(), TestRouter_ChainCallsInDeclaredOrder(), TestRouter_MatchedRouteServesNormally(), TestRouter_UnknownPathReturns404JSON() (+1 more)

### Community 34 - "Account Page (Web)"
Cohesion: 0.14
Nodes (4): bi(), add(), UserCabinet(), newAchievement()

### Community 35 - "Animation Recipes"
Cohesion: 0.15
Nodes (13): Animation Recipes Doc, Accordion/Collapse Recipe, Button Press Recipe, Crossfade Masking Recipe, Drawer/Sheet Recipe, Hold to Confirm Recipe, Modal Recipe, Scroll Reveal Recipe (+5 more)

### Community 36 - "Catalog Query Parsing"
Cohesion: 0.37
Nodes (12): parseBoolPtr(), parseFloatPtr(), parseIntPtr(), parseLat(), parseLimit(), parseListQuery(), parseLng(), parseSort() (+4 more)

### Community 37 - "JSON Decode Helpers"
Cohesion: 0.24
Nodes (11): decodeError(), DecodeJSON(), TestJSON_DecodeJSON_BodyTooLarge(), TestJSON_DecodeJSON_EmptyBody(), TestJSON_DecodeJSON_HappyPath(), TestJSON_DecodeJSON_MalformedJSON(), TestJSON_DecodeJSON_UnknownField(), TestJSON_WriteJSON_ContentType() (+3 more)

### Community 38 - "Geo Detection Utilities"
Cohesion: 0.17
Nodes (6): detectCoords(), detectRegionByGPS(), haversine(), nearestRegion(), useGPS(), locateMe()

### Community 39 - "Animation Skills Bundle"
Cohesion: 0.24
Nodes (12): Animate Skill, Find Animation Opportunities Skill, Animation Plan Template, Improve Animations Skill, Prototype Picker Spec (PICKER.md), Prototype Skill, Animate Build Sequence, Emil Kowalski Design Philosophy (+4 more)

### Community 40 - "Milestone 1 Catalog Tasks"
Cohesion: 0.18
Nodes (12): E3.4 — Optimistic Locking via ETag/If-Match, Milestone 1 — Catalog Read-Path (Tasks 17-31), Task 18 — Migration 00003 Concurrent Indexes, Task 23 — Keyset Pagination, Task 24 — GetByID via LATERAL json_agg, Task 28 — ETag Handler for Institution Card, Task 30 — cmd/devseed Seeder, golangci-lint v2 Config (+4 more)

### Community 41 - "Institution Get Handler & ETag"
Cohesion: 0.26
Nodes (9): buildETag(), GetHandler(), ListHandler(), classify(), errFields(), errMessage(), WriteError(), getService (+1 more)

### Community 42 - "Clock Package Tests"
Cohesion: 0.26
Nodes (9): Clock, New(), NewFake(), TestFake_AdvanceMovesNow(), TestFake_NowReturnsStart(), TestNew_ReturnsRealTime(), doRateLimitedRequest(), TestRateLimit_AllowsUpToLimitThenBlocks() (+1 more)

### Community 43 - "Healthz/Readyz Handler Tests"
Cohesion: 0.26
Nodes (10): Dependency, Healthz(), pingWithTimeout(), Readyz(), TestHealth_Healthz_ReturnsOKWithoutDependencies(), TestHealth_Readyz_FailingPingerReturns503WithName(), TestHealth_Readyz_MultipleDependenciesOnlyFailingListed(), TestHealth_Readyz_RespectsTimeoutEvenWhenPingerIgnoresContext() (+2 more)

### Community 44 - "DB Schema Ownership & Employment"
Cohesion: 0.20
Nodes (11): Schema Ownership per Service, auth.employment_claims, reviews.employer_reviews, EduHub Functional Requirements, FR-01/02/03 — Search, Facet Filters, Geo Search, FR-06 — Full Institution Card, FR-25 — Pricing Tiers (No Search Boost), FR-41 — Employer Reviews (+3 more)

### Community 45 - "Auth Schema Tables"
Cohesion: 0.22
Nodes (11): E2.6 — Children CRUD with PII Minimization, auth.children, auth.oauth_identities, auth.refresh_tokens, auth.users, auth.verification_codes, reviews.review_metrics, reviews.reviews (+3 more)

### Community 46 - "Backend Docker Compose Services"
Cohesion: 0.27
Nodes (11): api Service (hot-reload, air), db Service (Postgres+PostGIS), migrate Service, minio Service, redis Service, Backend Docker Compose Stack (db/redis/minio/migrate/api), Method 1: Full Docker Stack (hot-reload), Local Development Guide (+3 more)

### Community 47 - "Backend Config Loader"
Cohesion: 0.31
Nodes (9): Config, Load(), parseCORSAllowedOrigins(), TestLoad_AllVarsSet(), TestLoad_CORSAllowedOriginsParsesCommaSeparated(), TestLoad_CORSAllowedOriginsUnsetIsEmpty(), TestLoad_DefaultHTTPAddr(), TestLoad_MissingDatabaseURL() (+1 more)

### Community 48 - "i18n & Vacancy Filters"
Cohesion: 0.20
Nodes (4): useT(), Chip(), clearAll(), VacancyRow()

### Community 49 - "Animation Audit Categories"
Cohesion: 0.20
Nodes (10): Audit Category: Accessibility, Audit Category: Cohesion & Tokens, Audit Category: Easing & Duration, Audit Category: Interruptibility, Audit Category: Missed Opportunities, Audit Category: Performance, Audit Category: Physicality & Origin, Audit Category: Purpose & Frequency (+2 more)

### Community 50 - "Devseed Main & Integration Test"
Cohesion: 0.42
Nodes (8): main(), Seed(), countApprovedInstitutions(), countInstitutionSeedRefs(), testDatabaseURL(), TestSeed_FirstRun_Creates9ApprovedInstitutions(), TestSeed_NonDevAppEnv_RefusesWithoutTouchingDB(), TestSeed_SecondRun_IsIdempotent()

### Community 51 - "Config Loader Tests"
Cohesion: 0.31
Nodes (8): Config, Load(), parseCORSAllowedOrigins(), TestLoad_AllVarsSet(), TestLoad_CORSAllowedOriginsParsesCommaSeparated(), TestLoad_CORSAllowedOriginsUnsetIsEmpty(), TestLoad_DefaultHTTPAddr(), TestLoad_MissingDatabaseURL()

### Community 52 - "CLAUDE.md Rules & ADR-0001"
Cohesion: 0.29
Nodes (8): Chosen Variant C: Separation of Concerns, communications.vacancies, ADR-0001: Separate Vacancy Detail Page, Graphify Knowledge Graph Rules, Mandatory Behavior Rules, NEVER Rules, EduHub Project Contract (CLAUDE.md), FR-36/37 — Vacancy Board, Applications, Employer Outreach

### Community 53 - "Design Engineering Philosophy Skills"
Cohesion: 0.43
Nodes (8): Animation Vocabulary Glossary, Design Engineering (Emil Kowalski Philosophy), Awwwards-Level Design Engineering, Vanguard UI Architect, Premium Utilitarian Minimalism UI Architect, Redesign Existing Projects Skill, Reviewing Animations, Animation Standards Reference

### Community 54 - "Rate Limit Middleware"
Cohesion: 0.46
Nodes (7): clientIP(), RateLimit(), retryAfterSeconds(), writeRateLimitedError(), log/slog.Logger, net/http.Request, time.Duration

### Community 56 - "CORS Middleware"
Cohesion: 0.43
Nodes (5): CORS(), TestCORS_PreflightAllowedOriginReturns204WithHeaders(), TestCORS_PreflightDisallowedOriginNoAllowHeader(), TestCORS_RegularRequestAllowedOriginSetsHeaderAndCallsNext(), TestCORS_RegularRequestDisallowedOriginNoHeaderButCallsNext()

### Community 57 - "Panic Recovery Middleware"
Cohesion: 0.43
Nodes (5): Recover(), TestRecover_CatchesPanicAndReturns500(), TestRecover_NoPanicPassesThrough(), TestRecover_RequestIDInLog(), writeInternalServerError()

### Community 58 - "Site Layout & Footer"
Cohesion: 0.38
Nodes (3): EDUHUB_SOCIALS, Footer(), MaintenanceBanner()

### Community 59 - "Geo Locate Utility"
Cohesion: 0.38
Nodes (6): locateMe(), detectCoords(), detectRegionByGPS(), haversine(), nearestRegion(), REGION_CENTROIDS

### Community 60 - "SMM Rating Rules"
Cohesion: 0.33
Nodes (6): FR-14/29/30 — 8-Metric Rating, Threshold, Decay, Rule: 8 Rating Metrics, Not 9, Rule: /about Investor Page Permanently Banned from SMM, Channel Prioritization (Instagram/VK/Facebook/Telegram/TikTok), SMM Content Plan (EduHub Tajikistan), Rubric: Разбор одного критерия

### Community 63 - "Redis Cache Perf Baseline"
Cohesion: 0.50
Nodes (5): Redis Caching Strategy (catalog:list:v{version}, TTL 60s, singleflight), Task 31 — Redis Cache + Singleflight, Catalog Perf Baseline (Task 31), No-Cache Control Run (p95=7.9ms), Redis Cache Perf Result (p95=2.9ms)

### Community 65 - "SMM Site Audit Findings"
Cohesion: 0.40
Nodes (5): communications.visit_requests, Site Audit Findings (2026-08-21), Finding: Hardcoded Amenities & Fake Rating Histogram, Finding: Broken Hero Filter Chips, Finding: Broken Visit Request Form

### Community 66 - "Root Layout & Fonts"
Cohesion: 0.40
Nodes (3): inter, metadata, montserrat

### Community 67 - "Sonner Toast Skill"
Cohesion: 0.67
Nodes (4): Sonner API Reference (API.md), Ask Sonner Skill, Toast Recipe, Sonner

### Community 70 - "Web README Boilerplate"
Cohesion: 0.50
Nodes (3): next/font Geist optimization, create-next-app bootstrap, Deploy on Vercel

### Community 71 - "UI Design Recipes Misc"
Cohesion: 0.67
Nodes (3): Spatial Consistency, Base UI, Dropdown/Popover Recipe

### Community 72 - "Transactional Outbox Pattern"
Cohesion: 1.00
Nodes (3): Transactional Outbox + Redis Streams Pattern, communications.processed_events, reviews.outbox

## Knowledge Gaps
- **294 isolated node(s):** `Config`, `Dependency`, `Clock`, `create-next-app bootstrap`, `next/font Geist optimization` (+289 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **56 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `RequestID()` connect `Access Log Middleware` to `Devseed Demo Data`, `Institution Get Handler & ETag`, `Rate Limit Middleware`, `Panic Recovery Middleware`, `CORS Middleware Tests`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Why does `New()` connect `Health/Readyz Handler` to `HTTP Router & Chain`, `HTTP Error Taxonomy (apperr)`, `API Deps & Pinger Health Check`, `Clock Abstraction`, `Access Log Middleware Tests`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `main()` connect `HTTP Router & Chain` to `API Deps & Pinger Health Check`, `Health/Readyz Handler`, `Config Loader Tests`, `Access Log Middleware Tests`, `CORS Middleware Tests`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **What connects `Config`, `Dependency`, `Clock` to the rest of the system?**
  _294 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Catalog Repo & Transport Tests` be split into smaller, more focused modules?**
  _Cohesion score 0.06101409636019356 - nodes in this community are weakly interconnected._
- **Should `Business Plan Sections` be split into smaller, more focused modules?**
  _Cohesion score 0.050616050616050616 - nodes in this community are weakly interconnected._
- **Should `HTTP Error Taxonomy (apperr)` be split into smaller, more focused modules?**
  _Cohesion score 0.0815686274509804 - nodes in this community are weakly interconnected._