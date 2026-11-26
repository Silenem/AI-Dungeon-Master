package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// --- USER ---

type User struct {
	BaseEntity
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"` // Не отдаём в JSON на фронт
	Nickname     string  `json:"nickname"`
	AvatarURL    *string `json:"avatar_url"`
}

func NewUser(email, passwordHash, nickname string) *User { return nil }
func (u *User) CheckPassword(rawPassword string) bool    { return false }
func (u *User) UpdateProfile(nickname string, avatarURL *string) error { return nil }


// --- CHARACTER ---

type Attributes struct {
	Strength     int16 `json:"strength"`
	Dexterity    int16 `json:"dexterity"`
	Willpower    int16 `json:"willpower"`
	Constitution int16 `json:"constitution"`
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

func NewCharacter(userID uuid.UUID, name, classID string, stats Attributes, rulesetVer int) (*Character, error) {
	return nil, nil
}
func (c *Character) ValidateStats() error       { return nil }
func (c *Character) LevelUp() error             { return nil }
func (c *Character) SetSubclass(id string) error { return nil }


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

func NewSession(char *Character, scenario *Scenario) *Session { return nil }
func (s *Session) SyncFromActive(active *ActiveSession) error { return nil }
func (s *Session) AdvanceMemoryQueue(turnNo int64)            {}
func (s *Session) Finish(status string)                       {}