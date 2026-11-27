package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-dungeon-master/internal/domain"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	SaveActiveSession(ctx context.Context, session *domain.ActiveSession) error
	GetActiveSession(ctx context.Context, sessionID string) (*domain.ActiveSession, error)
	PushRecentTurn(ctx context.Context, sessionID string, turn *domain.Turn) error
	GetRecentTurns(ctx context.Context, sessionID string, count int) ([]*domain.Turn, error)
	PushPendingMemory(ctx context.Context, sessionID string, turn *domain.Turn) error
	ClearPendingMemory(ctx context.Context, sessionID string) error
}

type redisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string, db int) Store {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &redisStore{client: rdb}
}

func (s *redisStore) SaveActiveSession(ctx context.Context, session *domain.ActiveSession) error {
	key := fmt.Sprintf("session:%s:state", session.SessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (s *redisStore) GetActiveSession(ctx context.Context, sessionID string) (*domain.ActiveSession, error) {
	key := fmt.Sprintf("session:%s:state", sessionID)
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var session domain.ActiveSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *redisStore) PushRecentTurn(ctx context.Context, sessionID string, turn *domain.Turn) error {
	key := fmt.Sprintf("session:%s:recent-turns", sessionID)
	data, err := json.Marshal(turn)
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 19) // Оставляем только 20 последних
	_, err = pipe.Exec(ctx)
	return err
}

func (s *redisStore) GetRecentTurns(ctx context.Context, sessionID string, count int) ([]*domain.Turn, error) {
	key := fmt.Sprintf("session:%s:recent-turns", sessionID)
	results, err := s.client.LRange(ctx, key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	var turns []*domain.Turn
	for _, res := range results {
		var turn domain.Turn
		if err := json.Unmarshal([]byte(res), &turn); err == nil {
			turns = append(turns, &turn)
		}
	}
	return turns, nil
}

func (s *redisStore) PushPendingMemory(ctx context.Context, sessionID string, turn *domain.Turn) error {
	key := fmt.Sprintf("session:%s:pending-memory", sessionID)
	data, err := json.Marshal(turn)
	if err != nil {
		return err
	}
	return s.client.LPush(ctx, key, data).Err()
}

func (s *redisStore) ClearPendingMemory(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s:pending-memory", sessionID)
	return s.client.Del(ctx, key).Err()
}