CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE auth_provider AS ENUM ('local', 'google');
CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TYPE application_status AS ENUM (
    'wishlist', 'applied', 'screening', 'interview',
    'offer', 'accepted', 'rejected', 'withdrawn', 'ghosted'
);
CREATE TYPE work_arrangement AS ENUM ('remote', 'onsite', 'hybrid');
CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'internship', 'freelance');

CREATE TYPE event_type AS ENUM (
    'applied', 'phone_screen', 'interview', 'assessment',
    'offer', 'follow_up', 'deadline', 'note', 'status_changed'
);
CREATE TYPE email_token_type AS ENUM ('verify', 'reset');

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         CITEXT NOT NULL,
    password_hash TEXT,
    auth_provider auth_provider NOT NULL DEFAULT 'local',
    full_name     TEXT NOT NULL,
    avatar_url    TEXT,
    timezone      TEXT NOT NULL DEFAULT 'Asia/Jakarta',
    is_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned     BOOLEAN NOT NULL DEFAULT FALSE,
    banned_at     TIMESTAMPTZ,
    ban_reason    TEXT,
    role          user_role NOT NULL DEFAULT 'user',
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX users_email_unique ON users (email) WHERE deleted_at IS NULL;

CREATE TABLE job_applications (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users (id),
    company_name     TEXT NOT NULL,
    position_title   TEXT NOT NULL,
    job_url          TEXT,
    work_arrangement work_arrangement,
    employment_type  employment_type,
    location         TEXT,
    source           TEXT,
    status           application_status NOT NULL DEFAULT 'applied',
    applied_date     DATE,
    salary_min       NUMERIC(14, 2),
    salary_max       NUMERIC(14, 2),
    currency         TEXT,
    notes            TEXT,
    is_archived      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX job_applications_user_status_idx ON job_applications (user_id, status);
CREATE INDEX job_applications_user_applied_idx ON job_applications (user_id, applied_date, id);
CREATE INDEX job_applications_user_updated_idx ON job_applications (user_id, updated_at, id);

CREATE TABLE application_events (
    id             BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES job_applications (id),
    user_id        BIGINT NOT NULL REFERENCES users (id),
    type           event_type NOT NULL,
    title          TEXT NOT NULL,
    event_at       TIMESTAMPTZ NOT NULL,
    notes          TEXT,
    remind_at      TIMESTAMPTZ,
    reminded_at    TIMESTAMPTZ,
    status_from    application_status,
    status_to      application_status,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX application_events_app_event_idx ON application_events (application_id, event_at, id);
CREATE INDEX application_events_remind_idx ON application_events (remind_at) WHERE reminded_at IS NULL AND deleted_at IS NULL;

CREATE TABLE application_documents (
    id             BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES job_applications (id),
    user_id        BIGINT NOT NULL REFERENCES users (id),
    file_url       TEXT NOT NULL,
    file_name      TEXT NOT NULL,
    mime_type      TEXT NOT NULL,
    size_bytes     BIGINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX application_documents_app_idx ON application_documents (application_id) WHERE deleted_at IS NULL;

CREATE TABLE email_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id),
    type       email_token_type NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX email_tokens_user_type_idx ON email_tokens (user_id, type);
CREATE UNIQUE INDEX email_tokens_hash_idx ON email_tokens (token_hash);

CREATE TABLE refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id),
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX refresh_tokens_hash_idx ON refresh_tokens (token_hash);
CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id) WHERE revoked_at IS NULL;
