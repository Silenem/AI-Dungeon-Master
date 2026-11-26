CREATE TABLE users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname      TEXT NOT NULL,
    avatar_url    TEXT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE characters (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    class_id        TEXT NOT NULL,
    subclass_id     TEXT NULL,
    level           SMALLINT NOT NULL DEFAULT 1,
    strength        SMALLINT NOT NULL,
    dexterity       SMALLINT NOT NULL,
    willpower       SMALLINT NOT NULL,
    constitution    SMALLINT NOT NULL,
    starting_gear   JSONB NOT NULL DEFAULT '[]'::jsonb,
    ruleset_version INTEGER NOT NULL DEFAULT 1
);

