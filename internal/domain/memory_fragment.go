package domain

// MemoryFragment — документ для Elasticsearch (adventure-memory).
type MemoryFragment struct {
	ID              string   `json:"id"`
	SessionID       string   `json:"session_id"`
	ScenarioID      string   `json:"scenario_id"`
	ScenarioVersion int      `json:"scenario_version"`
	TurnNo          int64    `json:"turn_no"`
	GameDay         int      `json:"game_day"`
	Summary         string   `json:"summary"`
	EntityIDs       []string `json:"entity_ids"`
	EventIDs        []string `json:"event_ids"`
	PrevID          *string  `json:"prev_id"`     // Nullable (указатель)
	NextID          *string  `json:"next_id"`     // Nullable (указатель)
	RelatedIDs      []string `json:"related_ids"`
}

// LinkPrev — связывание с предыдущим фрагментом.
func (m *MemoryFragment) LinkPrev(prevID string) {
	m.PrevID = &prevID
}

// LinkNext — связывание со следующим фрагментом.
func (m *MemoryFragment) LinkNext(nextID string) {
	m.NextID = &nextID
}

// AddRelated — добавление смысловой связи.
func (m *MemoryFragment) AddRelated(relatedID string) {
	for _, id := range m.RelatedIDs {
		if id == relatedID {
			return // Предотвращаем дубликаты
		}
	}
	m.RelatedIDs = append(m.RelatedIDs, relatedID)
}