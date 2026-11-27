package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// --- КОНСТАНТЫ КЛАССОВ И ПОДКЛАССОВ (по Дереву классов v0) ---
const (
	ClassFighter = "fighter" // Боец
	ClassRanger  = "ranger"  // Следопыт
	ClassMystic  = "mystic"  // Мистик

	SubclassGuardian = "guardian" // Страж (Боец)
	SubclassSlayer   = "slayer"   // Убийца (Боец)
	SubclassRogue    = "rogue"    // Вор (Следопыт)
	SubclassHunter   = "hunter"   // Охотник (Следопыт)
	SubclassWizard   = "wizard"   // Волшебник (Мистик)
	SubclassSorcerer = "sorcerer" // Чародей (Мистик)
)

var allowedSubclasses = map[string][]string{
	ClassFighter: {SubclassGuardian, SubclassSlayer},
	ClassRanger:  {SubclassRogue, SubclassHunter},
	ClassMystic:  {SubclassWizard, SubclassSorcerer},
}

// --- USER ---

type User struct {
	BaseEntity
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"` // Не отдаём в JSON на фронт
	Nickname     string  `json:"nickname"`
	AvatarURL    *string `json:"avatar_url"`
}

func NewUser(email, passwordHash, nickname string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nickname = strings.TrimSpace(nickname)
	if email == "" || passwordHash == "" || nickname == "" {
		return nil, errors.New("email, passwordHash and nickname are required")
	}

	return &User{
		BaseEntity: BaseEntity{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	}, nil
}

// CheckPassword сверяет хеш пароля (позже можно заменить на bcrypt.CompareHashAndPassword).
func (u *User) CheckPassword(passwordHash string) bool {
	return u.PasswordHash != "" && u.PasswordHash == passwordHash
}

func (u *User) UpdateProfile(nickname string, avatarURL *string) error {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return errors.New("nickname cannot be empty")
	}
	u.Nickname = nickname
	u.AvatarURL = avatarURL
	return nil
}

// --- CHARACTER ---

type Attributes struct {
	Strength     int16 `json:"strength"`
	Dexterity    int16 `json:"dexterity"`
	Willpower    int16 `json:"willpower"`
	Constitution int16 `json:"constitution"`
}

// Modifier возвращает D&D-модификатор характеристики: (значение - 10) / 2.
func (a Attributes) Modifier(stat int16) int {
	diff := int(stat) - 10
	if diff < 0 && diff%2 != 0 {
		return diff/2 - 1
	}
	return diff / 2
}

type Character struct {
	BaseEntity
	UserID         uuid.UUID       `json:"user_id"`
	Name           string          `json:"name"`
	ClassID        string          `json:"class_id"`
	SubclassID     *string         `json:"subclass_id"`
	Level          int16           `json:"level"`
	Stats          Attributes      `json:"stats"`
	StartingGear   json.RawMessage `json:"starting_gear"` // jsonb
	RulesetVersion int             `json:"ruleset_version"`
}

func NewCharacter(userID uuid.UUID, name, classID string, rulesetVer int) (*Character, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("character name is required")
	}

	stats, gear, err := defaultClassPreset(classID)
	if err != nil {
		return nil, err
	}

	c := &Character{
		BaseEntity: BaseEntity{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		UserID:         userID,
		Name:           name,
		ClassID:        classID,
		Level:          1,
		Stats:          stats,
		StartingGear:   gear,
		RulesetVersion: rulesetVer,
	}

	if err := c.ValidateStats(); err != nil {
		return nil, err
	}
	return c, nil
}

// defaultClassPreset задаёт базу 10, +4 к основной характеристике, +2 к Телосложению и стартовый лут.
func defaultClassPreset(classID string) (Attributes, json.RawMessage, error) {
	switch classID {
	case ClassFighter:
		return Attributes{Strength: 14, Dexterity: 10, Willpower: 10, Constitution: 12},
			json.RawMessage(`["Чешуйчатый доспех", "Копьё"]`), nil
	case ClassRanger:
		return Attributes{Strength: 10, Dexterity: 14, Willpower: 10, Constitution: 12},
			json.RawMessage(`["Кожаный доспех", "Длинный лук и колчан с 20 стрелами", "Короткий меч"]`), nil
	case ClassMystic:
		return Attributes{Strength: 10, Dexterity: 10, Willpower: 14, Constitution: 12},
			json.RawMessage(`["Обычная одежда", "Боевой посох"]`), nil
	default:
		return Attributes{}, nil, fmt.Errorf("unknown class_id: %s", classID)
	}
}

func (c *Character) ValidateStats() error {
	for name, val := range map[string]int16{
		"strength":     c.Stats.Strength,
		"dexterity":    c.Stats.Dexterity,
		"willpower":    c.Stats.Willpower,
		"constitution": c.Stats.Constitution,
	} {
		if val < 1 || val > 30 {
			return fmt.Errorf("stat %s out of range (1..30): %d", name, val)
		}
	}
	return nil
}

func (c *Character) LevelUp() error {
	if c.Level >= 20 {
		return errors.New("character already at max level")
	}
	c.Level++
	return nil
}

func (c *Character) SetSubclass(subclassID string) error {
	allowed, ok := allowedSubclasses[c.ClassID]
	if !ok {
		return fmt.Errorf("unknown class_id: %s", c.ClassID)
	}
	for _, s := range allowed {
		if s == subclassID {
			c.SubclassID = &subclassID
			return nil
		}
	}
	return fmt.Errorf("subclass %s is not allowed for class %s", subclassID, c.ClassID)
}

// --- SESSION (долговременное сохранение в PostgreSQL) ---

type Session struct {
	BaseEntity
	CharacterID             uuid.UUID `json:"character_id"`
	ScenarioID              string    `json:"scenario_id"`
	ScenarioVersion         int       `json:"scenario_version"`
	RulesetVersion          int       `json:"ruleset_version"`
	Status                  string    `json:"status"`
	LastPlayedAt            time.Time `json:"last_played_at"`
	Revision                int64     `json:"revision"`
	MemoryQueuedThroughTurn int64     `json:"memory_queued_through_turn"`
	State                   Snapshot  `json:"state"` // jsonb
}

func NewSession(char *Character, scenario *Scenario, initialState Snapshot) *Session {
	now := time.Now().UTC()
	return &Session{
		BaseEntity: BaseEntity{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		CharacterID:     char.ID,
		ScenarioID:      scenario.ID,
		ScenarioVersion: scenario.Version,
		RulesetVersion:  char.RulesetVersion,
		Status:          "active",
		LastPlayedAt:    now,
		Revision:        0,
		State:           initialState,
	}
}

func (s *Session) SyncFromActive(active *ActiveSession) error {
	if active == nil {
		return errors.New("active session is nil")
	}
	if active.Revision < s.Revision {
		return fmt.Errorf("stale revision: active=%d < saved=%d", active.Revision, s.Revision)
	}
	s.Revision = active.Revision
	s.Status = active.Status
	s.LastPlayedAt = active.LastPlayedAt
	s.State = active.State
	return nil
}

func (s *Session) AdvanceMemoryQueue(turnNo int64) {
	if turnNo > s.MemoryQueuedThroughTurn {
		s.MemoryQueuedThroughTurn = turnNo
	}
}

func (s *Session) Finish(status string) {
	s.Status = status
	s.LastPlayedAt = time.Now().UTC()
}

// --- MEMORY OUTBOX ---

type MemoryOutboxEntry struct {
	SessionID     uuid.UUID `json:"session_id"`
	TurnNo        int64     `json:"turn_no"`
	Payload       Turn      `json:"payload"` // jsonb
	CreatedAt     time.Time `json:"created_at"`
	NextAttemptAt time.Time `json:"next_attempt_at"`
}

func NewMemoryOutboxEntry(sessionID uuid.UUID, turnNo int64, turn Turn) *MemoryOutboxEntry {
	now := time.Now().UTC()
	return &MemoryOutboxEntry{
		SessionID:     sessionID,
		TurnNo:        turnNo,
		Payload:       turn,
		CreatedAt:     now,
		NextAttemptAt: now,
	}
}

func (m *MemoryOutboxEntry) ScheduleRetry(delay time.Duration) {
	m.NextAttemptAt = time.Now().UTC().Add(delay)
}

// --- ИНТЕРФЕЙСЫ РЕПОЗИТОРИЕВ ---

type PostgresStorage interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	CreateCharacter(ctx context.Context, c *Character) error
	GetCharacterByID(ctx context.Context, id uuid.UUID) (*Character, error)

	SaveSessionWithOutbox(ctx context.Context, s *Session, entries []MemoryOutboxEntry) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*Session, error)

	FetchPendingOutbox(ctx context.Context, limit int) ([]MemoryOutboxEntry, error)
	MarkOutboxProcessed(ctx context.Context, sessionID uuid.UUID, turnNo int64) error
}
