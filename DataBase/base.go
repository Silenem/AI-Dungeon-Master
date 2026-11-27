package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity — базовая структура для всех сущностей с UUID и датой создания.
type BaseEntity struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// VersionedEntity — базовая структура для версионируемых сущностей (сценарии и т.д.).
type VersionedEntity struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}
