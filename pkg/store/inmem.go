package store

import (
	"fmt"
	"sync"

	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
)

type InMemoryStore struct {
	db sync.Map
}

func NewInMemStore() *InMemoryStore {
	return &InMemoryStore{db: sync.Map{}}
}

func (i *InMemoryStore) Get(id string) ([]api.Alert, error) {
	a, ok := i.db.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return i.alertFromValue(a)
}

func (i *InMemoryStore) Set(id string, alerts []api.Alert) error {
	i.db.Store(id, alerts)
	return nil
}

func (i *InMemoryStore) List() ([]api.MessageEntry, error) {
	var (
		contents []api.MessageEntry
		err      error
	)

	i.db.Range(func(key, value interface{}) bool {
		alerts, e := i.alertFromValue(value)
		if e != nil {
			err = e
			return false
		}

		id, e := i.idFromKey(key)
		if e != nil {
			err = e
			return false
		}

		contents = append(contents, api.MessageEntry{
			ID:     id,
			Alerts: alerts,
		})
		return true
	})
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (i *InMemoryStore) alertFromValue(from interface{}) ([]api.Alert, error) {
	alerts, ok := from.([]api.Alert)
	if !ok {
		return nil, fmt.Errorf("failed assertion from map value to list of alerts: %w", ErrInternal)
	}
	return alerts, nil
}

func (i *InMemoryStore) idFromKey(from interface{}) (string, error) {
	id, ok := from.(string)
	if !ok {
		return "", fmt.Errorf("failed assertion from map key to list of alerts: %w", ErrInternal)
	}
	return id, nil
}
