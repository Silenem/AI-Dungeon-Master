package domain

import "errors"

// GlobalEvent — глобальное событие сценария.
type GlobalEvent struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Conditions interface{} `json:"conditions"` // Условия срабатывания
	Effects    interface{} `json:"effects"`    // Эффекты при срабатывании
}

// CheckConditions — проверка условий (заглушка под бизнес-логику).
func (e *GlobalEvent) CheckConditions(session *ActiveSession) (bool, error) {
	if e.Conditions == nil {
		return false, errors.New("no conditions defined")
	}
	// Здесь в будущем будет логика проверки (например, game_day > X)
	return true, nil
}

// Scenario — описание сценария приключения.
type Scenario struct {
	ID              string        `json:"id"`
	Version         int           `json:"version"`
	Title           string        `json:"title"`
	Opening         string        `json:"opening"`
	InitialFacts    interface{}   `json:"initial_facts"`
	GlobalEvents    []GlobalEvent `json:"global_events"`
}