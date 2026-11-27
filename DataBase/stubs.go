package domain

import "time"

type Snapshot struct{}
type Turn struct{}

type Scenario struct {
	VersionedEntity
}

type ActiveSession struct {
	Revision     int64
	Status       string
	LastPlayedAt time.Time
	State        Snapshot
}
