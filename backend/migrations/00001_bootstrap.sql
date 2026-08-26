-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS reviews;
CREATE SCHEMA IF NOT EXISTS communications;
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS moderation;
CREATE SCHEMA IF NOT EXISTS platform;

-- +goose Down
DROP SCHEMA IF EXISTS platform CASCADE;
DROP SCHEMA IF EXISTS moderation CASCADE;
DROP SCHEMA IF EXISTS analytics CASCADE;
DROP SCHEMA IF EXISTS communications CASCADE;
DROP SCHEMA IF EXISTS reviews CASCADE;
DROP SCHEMA IF EXISTS catalog CASCADE;
DROP SCHEMA IF EXISTS auth CASCADE;

-- Расширения (postgis/pgcrypto/pg_trgm/unaccent) намеренно не дропаются:
-- это глобальный ресурс кластера, которым могут пользоваться другие схемы/БД,
-- откатывать его из-за одной миграции небезопасно.
