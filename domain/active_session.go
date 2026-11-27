package domain

import "time"

// ActiveSession — оперативное состояние сессии.
type ActiveSession struct {
	SessionID       string    `json:"session_id"`
	CharacterID     string    `json:"character_id"`
	Revision        int64     `json:"revision"`
	Status          string    `json:"status"`
	LastPlayedAt    time.Time `json:"last_played_at"` // При маршалинге даст RFC 3339
	State           Snapshot  `json:"state"`
	ActiveRequestID *string   `json:"active_request_id"` // nil, если нет активного запроса
}

// TryAcquireRequest — защита от параллельных запросов (идемпотентность).
func (s *ActiveSession) TryAcquireRequest(requestID string) bool {
	if s.ActiveRequestID != nil && *s.ActiveRequestID == requestID {
		// Запрос уже выполняется или выполнен
		return false
	}
	s.ActiveRequestID = &requestID
	return true
}

// ApplyTurn — применяет ход, обновляет ревизию и время.
func (s *ActiveSession) ApplyTurn(turn *Turn) {
	s.Revision++
	s.LastPlayedAt = time.Now().UTC()
	s.State.LastTurn = turn
	s.ActiveRequestID = nil // Сброс после успешного применения
}