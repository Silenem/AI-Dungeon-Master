package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Turn struct {
	RequestID   uuid.UUID       `json:"request_id"`
	TurnNo      int64           `json:"turn_no"`
	GameDay     int             `json:"game_day"`
	PlayerInput json.RawMessage `json:"player_input"`
	Result      json.RawMessage `json:"result"`
	Narration   *string         `json:"narration"`
}

type Snapshot struct {
	SchemaVersion int             `json:"schema_version"`
	Hero          json.RawMessage `json:"hero"`
	World         json.RawMessage `json:"world"`
	Mode          string          `json:"mode"`
	ZeroHPMode    string          `json:"zero_hp_mode"`
	Combat        json.RawMessage `json:"combat"`
	LastTurn      *Turn           `json:"last_turn"`
}

type ActiveSession struct {
	CharacterID     uuid.UUID  `json:"character_id"`
	Revision        int64      `json:"revision"`
	Status          string     `json:"status"`
	LastPlayedAt    time.Time  `json:"last_played_at"`
	State           Snapshot   `json:"state"`
	ActiveRequestID *uuid.UUID `json:"active_request_id"`
}

type Scenario struct {
	VersionedEntity
	Title        string            `json:"title"`
	Opening      string            `json:"opening"`
	InitialFacts map[string]bool   `json:"initial_facts"`
	GlobalEvents []json.RawMessage `json:"global_events"`
}
