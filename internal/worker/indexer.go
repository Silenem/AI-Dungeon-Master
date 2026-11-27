package worker

import (
	"context"
	"log"
	"time"

	"ai-dungeon-master/internal/domain"
	"ai-dungeon-master/internal/storage/elastic"
)

// OutboxTask — задача из очереди memory_outbox.
// Связывает идентификатор сессии и сам ход.
type OutboxTask struct {
	SessionID string
	Turn      *domain.Turn
}

// OutboxRepository — интерфейс для работы с memory_outbox (реализуется в PostgreSQL).
type OutboxRepository interface {
	// FetchUnprocessed забирает необработанные задачи (например, пачками по limit).
	FetchUnprocessed(ctx context.Context, limit int) ([]*OutboxTask, error)
	
	// MarkProcessed помечает задачу как успешно обработанную.
	MarkProcessed(ctx context.Context, sessionID string, turnNo int64) error
}

// IndexerWorker — фоновый воркер индексации памяти.
type IndexerWorker struct {
	outbox  OutboxRepository
	esStore elastic.SearchStore
}

func NewIndexerWorker(outbox OutboxRepository, esStore elastic.SearchStore) *IndexerWorker {
	return &IndexerWorker{
		outbox:  outbox,
		esStore: esStore,
	}
}

// Run — запуск воркера в горутине.
func (w *IndexerWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Indexer worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *IndexerWorker) processBatch(ctx context.Context) {
	tasks, err := w.outbox.FetchUnprocessed(ctx, 10) // Забираем пачками по 10
	if err != nil {
		log.Printf("Failed to fetch outbox: %v", err)
		return
	}

	for _, task := range tasks {
		if task == nil || task.Turn == nil {
			continue
		}

		turn := task.Turn

		// Безопасное извлечение текста: проверка на nil, чтобы избежать паники,
		// если ИИ не вернул narration (например, ход только с броском кубика).
		summary := ""
		if turn.Narration != nil {
			summary = *turn.Narration
		}

		// 1. Формируем фрагмент памяти для Elasticsearch.
		fragment := &domain.MemoryFragment{
			ID:        turn.RequestID, // или генерируем UUID
			SessionID: task.SessionID,
			TurnNo:    turn.TurnNo,
			GameDay:   turn.GameDay,
			Summary:   summary,
			// EntityIDs и EventIDs заполняются из результата ИИ
		}

		// 2. Индексируем в Elasticsearch.
		if err := w.esStore.IndexFragment(ctx, fragment); err != nil {
			log.Printf("Failed to index fragment %s: %v", fragment.ID, err)
			continue // Оставляем в outbox для ретрая
		}

		// 3. Обновляем связи (с предыдущим ходом).
		// В идеале нужно достать prev_id из ES или Redis.
		// ... (логика связывания) ...
		// w.esStore.UpdateLinks(ctx, fragment)

		// 4. Помечаем в БД как обработанное.
		if err := w.outbox.MarkProcessed(ctx, task.SessionID, turn.TurnNo); err != nil {
			log.Printf("Failed to mark processed turn %d for session %s: %v", turn.TurnNo, task.SessionID, err)
		}
	}
}