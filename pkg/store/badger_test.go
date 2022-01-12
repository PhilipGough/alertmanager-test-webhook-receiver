package store

import (
	"reflect"
	"testing"

	"github.com/go-kit/log"
	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
)

func TestKeyValueStore_Set(t *testing.T) {
	store, err := NewKeyValueStore("", log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Set("any", getTestAlerts()); err != nil {
		t.Fatal(err)
	}
}

func TestKeyValueStore_Get(t *testing.T) {
	store, err := NewKeyValueStore("", log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}

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

func TestNewKeyValueStore_List(t *testing.T) {
	store, err := NewKeyValueStore("", log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Set("any", getTestAlerts()); err != nil {
		t.Fatal(err)
	}
	result, err := store.List()
	if err != nil {
		t.Fatalf("expected retrieve to succeed but got %v", err)
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
