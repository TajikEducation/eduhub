-- +goose Up
-- Веха 2 (E2.1), ядро: только то, что нужно для register/login/JWT/RBAC.
-- oauth_identities/verification_codes/children — отложены до отдельных задач
-- (Google OAuth, email-verification, Children CRUD).
CREATE TABLE auth.users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL,
    password_hash TEXT,
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'institution', 'moderator', 'admin')),
    -- 'active' по умолчанию, а не 'unverified' из полного SRS-спека: email-verification
    -- ещё не реализован (отдельная задача), блокировать вход несуществующим шагом нельзя.
    status        TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('unverified', 'active', 'banned', 'deleted')),
    display_name  TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Регистронезависимая уникальность email — сравнение по lower(), не отдельная колонка/CITEXT.
CREATE UNIQUE INDEX users_email_lower_idx ON auth.users (lower(email));

-- E2.3: хранится только хеш refresh-токена, никогда сам токен. family_id зарезервирован
-- под будущий reuse-detection (отзыв всей семьи при повторном предъявлении использованного
-- токена) — в этой версии usecase его не проверяет, только пишет.
CREATE TABLE auth.refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL,
    family_id   UUID NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    replaced_by UUID REFERENCES auth.refresh_tokens(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX refresh_tokens_token_hash_idx ON auth.refresh_tokens (token_hash);
CREATE INDEX refresh_tokens_user_id_idx ON auth.refresh_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.users;
