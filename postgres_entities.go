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

