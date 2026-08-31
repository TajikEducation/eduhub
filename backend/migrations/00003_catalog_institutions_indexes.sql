-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY idx_institutions_geo
    ON catalog.institutions USING GIST (geo);

CREATE INDEX CONCURRENTLY idx_institutions_name_jsonb
    ON catalog.institutions USING GIN (name jsonb_path_ops);

CREATE INDEX CONCURRENTLY idx_institutions_name_trgm_ru
    ON catalog.institutions USING GIN ((name ->> 'ru') gin_trgm_ops);

CREATE INDEX CONCURRENTLY idx_institutions_name_trgm_tg
    ON catalog.institutions USING GIN ((name ->> 'tg') gin_trgm_ops);

CREATE INDEX CONCURRENTLY idx_institutions_types
    ON catalog.institutions USING GIN (types);

CREATE INDEX CONCURRENTLY idx_institutions_curriculum
    ON catalog.institutions USING GIN (curriculum);

CREATE INDEX CONCURRENTLY idx_institutions_program_level
    ON catalog.institutions USING GIN (program_level);

CREATE INDEX CONCURRENTLY idx_institutions_region_district
    ON catalog.institutions USING BTREE (region, district);

CREATE INDEX CONCURRENTLY idx_institutions_rating_approved
    ON catalog.institutions USING BTREE (rating_avg DESC, id)
    WHERE moderation_status = 'approved';

CREATE INDEX CONCURRENTLY idx_institutions_price_approved
    ON catalog.institutions USING BTREE (price)
    WHERE moderation_status = 'approved';

CREATE UNIQUE INDEX CONCURRENTLY idx_institutions_name_region_district_unique
    ON catalog.institutions (lower(name ->> 'ru'), region, district)
    WHERE moderation_status <> 'rejected';

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_name_region_district_unique;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_price_approved;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_rating_approved;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_region_district;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_program_level;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_curriculum;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_types;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_name_trgm_tg;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_name_trgm_ru;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_name_jsonb;
DROP INDEX CONCURRENTLY IF EXISTS catalog.idx_institutions_geo;
