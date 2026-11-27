package domain

// Combat — состояние боя.
type Combat struct {
	Round         int    `json:"round"`
	CurrentTurnID string `json:"current_turn_id"`
}

// Snapshot — состояние героя и мира.
type Snapshot struct {
	SchemaVersion int         `json:"schema_version"`
	Hero          interface{} `json:"hero"`   // В реальности структура Character
	World         interface{} `json:"world"`  // В реальности структура World
	Mode          string      `json:"mode"`   // "normal" или "combat"
	ZeroHPMode    string      `json:"zero_hp_mode"`
	Combat        *Combat     `json:"combat"` // nil, если не в бою (nullable)
	LastTurn      *Turn       `json:"last_turn"`
}

// EnterCombat — переход в боевой режим.
func (s *Snapshot) EnterCombat() {
	s.Mode = "combat"
	s.Combat = &Combat{Round: 1}
}

// ExitCombat — выход из боя.
func (s *Snapshot) ExitCombat() {
	s.Mode = "normal"
	s.Combat = nil
}