package store

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/dgraph-io/badger/v3"
	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
)

type KeyValueStore struct {
	db *badger.DB
}

// NewKeyValueStore creates a new Store at the provided path
// If path is empty an in-memory database is used
func NewKeyValueStore(path string, logger log.Logger) (*KeyValueStore, error) {
	var (
		db  *badger.DB
		err error
	)
	opts := badger.DefaultOptions(path)
	opts.Logger = &wrappedLogger{logger: logger}

	if path != "" {
		db, err = badger.Open(opts)
	} else {
		db, err = badger.Open(opts.WithInMemory(true))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return &KeyValueStore{db: db}, nil
}

func (k *KeyValueStore) Get(id string) ([]api.Alert, error) {
	var out []api.Alert
	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err == badger.ErrKeyNotFound {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &out); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, err
}

func (k *KeyValueStore) Set(id string, alerts []api.Alert) error {
	b, err := json.Marshal(alerts)
	if err != nil {
		return err
	}
	return k.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(id), b)
		return err
	})
}

func (k *KeyValueStore) List() ([]api.MessageEntry, error) {
	var entries []api.MessageEntry
	err := k.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var alerts []api.Alert
			item := it.Item()

			err := item.Value(func(v []byte) error {
				if err := json.Unmarshal(v, &alerts); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			entries = append(entries, api.MessageEntry{
				ID:     string(item.Key()),
				Alerts: alerts,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (k *KeyValueStore) toMessageEntry(key, v []byte) (*api.MessageEntry, error) {
	var alerts []api.Alert
	if err := json.Unmarshal(v, &alerts); err != nil {
		return nil, err
	}

	var id string
	if err := json.Unmarshal(key, &id); err != nil {
		return nil, err
	}

	return &api.MessageEntry{
		ID:     id,
		Alerts: alerts,
	}, nil
}

type wrappedLogger struct {
	logger log.Logger
}

func (l *wrappedLogger) Errorf(f string, v ...interface{}) {
	level.Error(l.logger).Log("msg", fmt.Sprintf(f, v...))
}

func (l *wrappedLogger) Warningf(f string, v ...interface{}) {
	level.Warn(l.logger).Log("msg", fmt.Sprintf(f, v...))
}

func (l *wrappedLogger) Infof(f string, v ...interface{}) {
	level.Info(l.logger).Log("msg", fmt.Sprintf(f, v...))
}

func (l *wrappedLogger) Debugf(f string, v ...interface{}) {
	level.Debug(l.logger).Log("msg", fmt.Sprintf(f, v...))
}
