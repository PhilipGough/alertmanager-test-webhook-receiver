package store

import "github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"

const (
	ErrNotFound = Error("not found")
	ErrInternal = Error("internal")
)

type Store interface {
	Get(id string) ([]api.Alert, error)
	Set(id string, alerts []api.Alert) error
	List() ([]api.MessageEntry, error)
}

type Error string

func (e Error) Error() string { return string(e) }
