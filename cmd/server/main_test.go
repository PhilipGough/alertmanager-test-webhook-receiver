package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
)

func TestGetHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/history/test_id", nil)
	if err != nil {
		t.Fatal(err)
	}

	srv := &server{
		router: mux.NewRouter(),
		logger: log.NewNopLogger(),
		store: mockStore{
			getFn: func(id string) ([]api.Alert, error) {
				if id != "test_id" {
					t.Fatalf("expected test_id but got %s", id)
				}
				return []api.Alert{
					{
						Status:      "test",
						Fingerprint: "test",
					},
				}, nil
			}},
	}
	srv.routes()

	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("expected 200 response")
	}
	expect := `[{"status":"test","labels":null,"annotations":null,"startsAt":"0001-01-01T00:00:00Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"test"}]`
	if strings.TrimSpace(w.Body.String()) != expect {
		t.Fatal("unexpected json response")
	}
}

func TestWebhookHandler(t *testing.T) {
	body, err := os.Open("testdata/request.json")
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/webhook", body)
	if err != nil {
		t.Fatal(err)
	}

	srv := &server{
		router: mux.NewRouter(),
		logger: log.NewNopLogger(),
		store: mockStore{
			setFn: func(id string, alerts []api.Alert) error {
				expect := "Test_webhook"
				if id != expect {
					t.Fatalf("wanted %s but got %s", expect, id)
				}

				start := `2018-08-03T09:52:26.739266876+02:00`
				end := `0001-01-01T00:00:00Z`

				tStart, err := time.Parse(time.RFC3339, start)
				if err != nil {
					t.Fatal(err)
				}

				tEnd, err := time.Parse(time.RFC3339, end)
				if err != nil {
					t.Fatal(err)
				}

				expectAlerts := []api.Alert{
					{
						Status: "firing",
						Labels: map[string]string{
							"alertname": "Test",
							"dc":        "eu-west-1",
							"instance":  "localhost:9090",
							"job":       "prometheus24",
						},
						Annotations: map[string]string{
							"description": "some description",
						},
						StartsAt:     tStart,
						EndsAt:       tEnd,
						GeneratorURL: "http://example.com",
					},
				}

				if !reflect.DeepEqual(expectAlerts, alerts) {
					t.Fatalf("unexpected alerts got %v wanted %v", alerts, expectAlerts)
				}
				return nil
			},
		},
		idGenerator: buildIdGenerator(defaultStoreIDTemplate),
	}
	srv.routes()

	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("expected 200 response")
	}
	expect := `{"id":"Test_webhook"}`
	if strings.TrimSpace(w.Body.String()) != expect {
		t.Fatal("unexpected json response")
	}
}

func TestListHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/history", nil)
	if err != nil {
		t.Fatal(err)
	}

	srv := &server{
		router: mux.NewRouter(),
		logger: log.NewNopLogger(),
		store: mockStore{
			listFn: func() ([]api.MessageEntry, error) {
				start := `2018-08-03T09:52:26.739266876+02:00`
				end := `0001-01-01T00:00:00Z`

				tStart, err := time.Parse(time.RFC3339, start)
				if err != nil {
					t.Fatal(err)
				}

				tEnd, err := time.Parse(time.RFC3339, end)
				if err != nil {
					t.Fatal(err)
				}

				return []api.MessageEntry{
					{
						ID: "1",
						Alerts: []api.Alert{
							{
								Status: "firing",
								Labels: map[string]string{
									"alertname": "Test",
									"dc":        "eu-west-1",
									"instance":  "localhost:9090",
									"job":       "prometheus24",
								},
								Annotations: map[string]string{
									"description": "some description",
								},
								StartsAt:     tStart,
								EndsAt:       tEnd,
								GeneratorURL: "http://example.com",
							},
						},
					},
				}, nil
			},
		},
		idGenerator: buildIdGenerator(defaultStoreIDTemplate),
	}
	srv.routes()

	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("expected 200 response")
	}
	expect := `[{"id":"1","alerts":[{"status":"firing","labels":{"alertname":"Test","dc":"eu-west-1","instance":"localhost:9090","job":"prometheus24"},"annotations":{"description":"some description"},"startsAt":"2018-08-03T09:52:26.739266876+02:00","endsAt":"0001-01-01T00:00:00Z","generatorURL":"http://example.com"}]}]`
	if strings.TrimSpace(w.Body.String()) != expect {
		t.Fatalf("unexpected json response %s", w.Body.String())
	}
}

func TestGeneratedIDFromPayload(t *testing.T) {
	samplePayload := getSamplePayload(t)
	t.Cleanup(func() {
		_ = samplePayload.Close()
	})

	var reqBody api.Message
	if err := json.NewDecoder(samplePayload).Decode(&reqBody); err != nil {
		t.Fatal(err)
	}

	testDefaultTmplFn := buildIdGenerator(defaultStoreIDTemplate)
	got, err := testDefaultTmplFn(reqBody)
	if err != nil {
		t.Fatal("expected default template to parse correctly")
	}
	expect := "Test_webhook"
	if got != expect {
		t.Fatalf("wanted %s but got %s", expect, got)
	}

	testCustomTmplFn := buildIdGenerator(`{{ .Version }}-{{ .Status }}`)
	got, err = testCustomTmplFn(reqBody)
	if err != nil {
		t.Fatal("expected custom valid template to parse correctly")
	}
	expect = "4-firing"
	if got != expect {
		t.Fatalf("wanted %s but got %s", expect, got)
	}
}

func getSamplePayload(t *testing.T) io.ReadCloser {
	t.Helper()
	f, err := os.Open("testdata/request.json")
	if err != nil {
		t.Fatal(err)
	}
	return f
}

type mockStore struct {
	getFn  func(id string) ([]api.Alert, error)
	setFn  func(id string, alerts []api.Alert) error
	listFn func() ([]api.MessageEntry, error)
}

func (m mockStore) Get(id string) ([]api.Alert, error) {
	return m.getFn(id)
}

func (m mockStore) Set(id string, alerts []api.Alert) error {
	return m.setFn(id, alerts)
}

func (m mockStore) List() ([]api.MessageEntry, error) {
	return m.listFn()
}
