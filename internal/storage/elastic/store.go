package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-dungeon-master/internal/domain"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const IndexName = "adventure-memory"

type SearchStore interface {
	EnsureIndex(ctx context.Context) error
	IndexFragment(ctx context.Context, fragment *domain.MemoryFragment) error
	UpdateLinks(ctx context.Context, fragment *domain.MemoryFragment) error
	SearchMemory(ctx context.Context, sessionID, query string, limit int) ([]*domain.MemoryFragment, error)
}

type elasticStore struct {
	client *elasticsearch.Client
}

func NewElasticStore(addresses []string) (SearchStore, error) {
	cfg := elasticsearch.Config{Addresses: addresses}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &elasticStore{client: es}, nil
}

func (s *elasticStore) EnsureIndex(ctx context.Context) error {
	res, err := s.client.Indices.Exists([]string{IndexName})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil // Индекс уже есть
	}

	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"session_id": {"type": "keyword"},
				"scenario_id": {"type": "keyword"},
				"scenario_version": {"type": "integer"},
				"turn_no": {"type": "long"},
				"game_day": {"type": "integer"},
				"summary": {"type": "text"},
				"entity_ids": {"type": "keyword"},
				"event_ids": {"type": "keyword"},
				"prev_id": {"type": "keyword"},
				"next_id": {"type": "keyword"},
				"related_ids": {"type": "keyword"}
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: IndexName,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("failed to create index: %s", createRes.String())
	}
	return nil
}

func (s *elasticStore) IndexFragment(ctx context.Context, fragment *domain.MemoryFragment) error {
	data, err := json.Marshal(fragment)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      IndexName,
		DocumentID: fragment.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true", 
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}
	return nil
}

func (s *elasticStore) UpdateLinks(ctx context.Context, fragment *domain.MemoryFragment) error {
	updateBody := map[string]interface{}{
		"doc": map[string]interface{}{
			"prev_id":     fragment.PrevID,
			"next_id":     fragment.NextID,
			"related_ids": fragment.RelatedIDs,
		},
	}
	data, _ := json.Marshal(updateBody)

	req := esapi.UpdateRequest{
		Index:      IndexName,
		DocumentID: fragment.ID,
		Body:       bytes.NewReader(data),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *elasticStore) SearchMemory(ctx context.Context, sessionID, query string, limit int) ([]*domain.MemoryFragment, error) {
	// Используем filter для session_id, чтобы поиск шел строго внутри сессии
	queryBody := fmt.Sprintf(`{
		"query": {
			"bool": {
				"filter": [
					{"term": {"session_id": "%s"}}
				],
				"must": [
					{"match": {"summary": "%s"}}
				]
			}
		},
		"size": %d
	}`, sessionID, query, limit)

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(IndexName),
		s.client.Search.WithBody(strings.NewReader(queryBody)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	var fragments []*domain.MemoryFragment
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		sourceData, _ := json.Marshal(source)
		var frag domain.MemoryFragment
		json.Unmarshal(sourceData, &frag)
		fragments = append(fragments, &frag)
	}

	return fragments, nil
}