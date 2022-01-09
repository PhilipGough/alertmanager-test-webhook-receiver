package store

import (
	"reflect"
	"testing"

	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
)

func TestInMemoryStore_Set(t *testing.T) {
	store := NewInMemStore()
	if err := store.Set("any", getTestAlerts()); err != nil {
		t.Fatal(err)
	}
}

func TestInMemoryStore_Get(t *testing.T) {
	store := NewInMemStore()
	if err := store.Set("any", getTestAlerts()); err != nil {
		t.Fatal(err)
	}

	result, err := store.Get("any")
	if err != nil {
		t.Fatal("expected retrieve to succeed")
	}

	if !reflect.DeepEqual(result, getTestAlerts()) {
		t.Fatalf("wanted %v got %v", getTestAlerts(), result)
	}
}

func TestInMemoryStore_List(t *testing.T) {
	store := NewInMemStore()
	if err := store.Set("any", getTestAlerts()); err != nil {
		t.Fatal(err)
	}

	result, err := store.List()
	if err != nil {
		t.Fatal("expected retrieve to succeed")
	}

	expect := []api.MessageEntry{
		{
			ID:     "any",
			Alerts: getTestAlerts()},
	}

	if !reflect.DeepEqual(result, expect) {
		t.Fatalf("wanted %v got %v", expect, result)
	}
}

func getTestAlerts() []api.Alert {
	return []api.Alert{
		{
			Status:       "firing",
			GeneratorURL: "https://test.com",
		},
		{
			Status:       "pending",
			GeneratorURL: "https://example.com",
		},
	}
}
