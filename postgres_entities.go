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