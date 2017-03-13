package eventsource

import (
	"context"
	"sync"
)

type Store interface {
	Save(ctx context.Context, serializer Serializer, events ...interface{}) error
	Fetch(ctx context.Context, serializer Serializer, aggregateID string, version int) ([]interface{}, int, error)
}

type memoryStore struct {
	mux        *sync.Mutex
	aggregates map[string][]EventMeta
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		mux:        &sync.Mutex{},
		aggregates: map[string][]EventMeta{},
	}
}

func (m *memoryStore) Save(ctx context.Context, serializer Serializer, events ...interface{}) error {
	for _, event := range events {
		meta, err := Inspect(event)
		if err != nil {
			return err
		}

		v, ok := m.aggregates[meta.AggregateID]
		if !ok {
			v = make([]EventMeta, 0, len(events))
			m.aggregates[meta.AggregateID] = v
		}

		m.aggregates[meta.AggregateID] = append(v, meta)
	}

	return nil
}

func (m *memoryStore) Fetch(ctx context.Context, serializer Serializer, aggregateID string, version int) ([]interface{}, int, error) {
	v, ok := m.aggregates[aggregateID]
	if !ok {
		return nil, 0, ErrNotFound
	}

	found := 0

	events := make([]interface{}, 0, len(v))
	for _, meta := range v {
		if meta.Version > version {
			break
		}
		events = append(events, meta.Event)
		found = meta.Version
	}

	return events, found, nil
}
