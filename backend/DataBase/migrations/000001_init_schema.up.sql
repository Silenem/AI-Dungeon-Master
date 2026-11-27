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


CREATE TABLE sessions (
    id                         UUID PRIMARY KEY,
    character_id               UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    scenario_id                TEXT NOT NULL,
    scenario_version           INTEGER NOT NULL,
    ruleset_version            INTEGER NOT NULL,
    status                     TEXT NOT NULL,
    last_played_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revision                   BIGINT NOT NULL DEFAULT 0,
    memory_queued_through_turn BIGINT NOT NULL DEFAULT 0,
    state                      JSONB NOT NULL
);

CREATE TABLE memory_outbox (
    session_id      UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    turn_no         BIGINT NOT NULL,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, turn_no)
);

-- Индексы для быстрого поиска по внешним ключам и очереди воркера
CREATE INDEX idx_characters_user_id ON characters(user_id);
CREATE INDEX idx_sessions_character_id ON sessions(character_id);
CREATE INDEX idx_memory_outbox_next_attempt ON memory_outbox(next_attempt_at);