package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "ai-dungeon-master"
)

var ErrNotFound = errors.New("record not found")

type Storage struct {
	pool *pgxpool.Pool
}

// Проверка на этапе компиляции, что Storage реализует интерфейс domain.PostgresStorage.
var _ domain.PostgresStorage = (*Storage)(nil)

// New создаёт пул соединений и проверяет доступность PostgreSQL.
func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

// --- USERS ---

func (s *Storage) CreateUser(ctx context.Context, u *domain.User) error {
	const query = `
		INSERT INTO users (id, email, password_hash, nickname, avatar_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, query,
		u.ID, u.Email, u.PasswordHash, u.Nickname, u.AvatarURL, u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, nickname, avatar_url, created_at
		FROM users
		WHERE email = $1`

	var u domain.User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.AvatarURL, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// --- CHARACTERS ---

func (s *Storage) CreateCharacter(ctx context.Context, c *domain.Character) error {
	const query = `
		INSERT INTO characters (
			id, user_id, name, class_id, subclass_id, level,
			strength, dexterity, willpower, constitution,
			starting_gear, ruleset_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	gear := c.StartingGear
	if len(gear) == 0 {
		gear = json.RawMessage("[]")
	}

	_, err := s.pool.Exec(ctx, query,
		c.ID, c.UserID, c.Name, c.ClassID, c.SubclassID, c.Level,
		c.Stats.Strength, c.Stats.Dexterity, c.Stats.Willpower, c.Stats.Constitution,
		gear, c.RulesetVersion,
	)
	if err != nil {
		return fmt.Errorf("create character: %w", err)
	}
	return nil
}

func (s *Storage) GetCharacterByID(ctx context.Context, id uuid.UUID) (*domain.Character, error) {
	const query = `
		SELECT
			id, user_id, name, class_id, subclass_id, level,
			strength, dexterity, willpower, constitution,
			starting_gear, ruleset_version
		FROM characters
		WHERE id = $1`

	var c domain.Character
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Name, &c.ClassID, &c.SubclassID, &c.Level,
		&c.Stats.Strength, &c.Stats.Dexterity, &c.Stats.Willpower, &c.Stats.Constitution,
		&c.StartingGear, &c.RulesetVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get character by id: %w", err)
	}
	return &c, nil
}

// --- SESSIONS & MEMORY OUTBOX ---

// SaveSessionWithOutbox атомарно сохраняет сессию и добавляет новые задачи в memory_outbox.
func (s *Storage) SaveSessionWithOutbox(ctx context.Context, sess *domain.Session, entries []domain.MemoryOutboxEntry) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	stateJSON, err := json.Marshal(sess.State)
	if err != nil {
		return fmt.Errorf("marshal session state: %w", err)
	}

	const upsertSession = `
		INSERT INTO sessions (
			id, character_id, scenario_id, scenario_version, ruleset_version,
			status, last_played_at, revision, memory_queued_through_turn, state
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			last_played_at = EXCLUDED.last_played_at,
			revision = EXCLUDED.revision,
			memory_queued_through_turn = EXCLUDED.memory_queued_through_turn,
			state = EXCLUDED.state
		WHERE sessions.revision <= EXCLUDED.revision`

	_, err = tx.Exec(ctx, upsertSession,
		sess.ID, sess.CharacterID, sess.ScenarioID, sess.ScenarioVersion, sess.RulesetVersion,
		sess.Status, sess.LastPlayedAt, sess.Revision, sess.MemoryQueuedThroughTurn, stateJSON,
	)
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}

	const insertOutbox = `
		INSERT INTO memory_outbox (session_id, turn_no, payload, created_at, next_attempt_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, turn_no) DO NOTHING`

	for _, entry := range entries {
		payloadJSON, err := json.Marshal(entry.Payload)
		if err != nil {
			return fmt.Errorf("marshal outbox payload (turn %d): %w", entry.TurnNo, err)
		}

		_, err = tx.Exec(ctx, insertOutbox,
			entry.SessionID, entry.TurnNo, payloadJSON, entry.CreatedAt, entry.NextAttemptAt,
		)
		if err != nil {
			return fmt.Errorf("insert outbox (turn %d): %w", entry.TurnNo, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Storage) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	const query = `
		SELECT
			id, character_id, scenario_id, scenario_version, ruleset_version,
			status, last_played_at, revision, memory_queued_through_turn, state
		FROM sessions
		WHERE id = $1`

	var (
		sess      domain.Session
		stateJSON []byte
	)

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&sess.ID, &sess.CharacterID, &sess.ScenarioID, &sess.ScenarioVersion, &sess.RulesetVersion,
		&sess.Status, &sess.LastPlayedAt, &sess.Revision, &sess.MemoryQueuedThroughTurn, &stateJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}

	if err := json.Unmarshal(stateJSON, &sess.State); err != nil {
		return nil, fmt.Errorf("unmarshal session state: %w", err)
	}
	return &sess, nil
}

func (s *Storage) FetchPendingOutbox(ctx context.Context, limit int) ([]domain.MemoryOutboxEntry, error) {
	const query = `
		SELECT session_id, turn_no, payload, created_at, next_attempt_at
		FROM memory_outbox
		WHERE next_attempt_at <= NOW()
		ORDER BY next_attempt_at ASC, turn_no ASC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch pending outbox: %w", err)
	}
	defer rows.Close()

	var result []domain.MemoryOutboxEntry
	for rows.Next() {
		var (
			entry       domain.MemoryOutboxEntry
			payloadJSON []byte
		)
		if err := rows.Scan(
			&entry.SessionID, &entry.TurnNo, &payloadJSON, &entry.CreatedAt, &entry.NextAttemptAt,
		); err != nil {
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}
		if err := json.Unmarshal(payloadJSON, &entry.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal outbox payload: %w", err)
		}
		result = append(result, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox rows: %w", err)
	}
	return result, nil
}

func (s *Storage) MarkOutboxProcessed(ctx context.Context, sessionID uuid.UUID, turnNo int64) error {
	const query = `
		DELETE FROM memory_outbox
		WHERE session_id = $1 AND turn_no = $2`

	_, err := s.pool.Exec(ctx, query, sessionID, turnNo)
	if err != nil {
		return fmt.Errorf("mark outbox processed: %w", err)
	}
	return nil
}
