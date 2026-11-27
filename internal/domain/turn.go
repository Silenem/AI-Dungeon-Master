package domain

// Turn — игровой ход.
type Turn struct {
	RequestID   string      `json:"request_id"`
	TurnNo      int64       `json:"turn_no"`
	GameDay     int         `json:"game_day"`
	PlayerInput string      `json:"player_input"`
	Result      interface{} `json:"result"` // Результат бросков/действий (JSONB в PG)
	Narration   *string     `json:"narration,omitempty"` // Текст от ИИ (nullable)
}

// NewTurn — конструктор хода по player_input.
func NewTurn(reqID string, turnNo int64, day int, input string) *Turn {
	return &Turn{
		RequestID:   reqID,
		TurnNo:      turnNo,
		GameDay:     day,
		PlayerInput: input,
	}
}