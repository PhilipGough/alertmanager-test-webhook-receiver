package api

// Types are duplicated from https://github.com/prometheus/alertmanager
// in order to avoid the dependency
import (
	"encoding/json"
	"fmt"
	"time"
)

// Message is the POST request body for the webhook and maps to
// https://github.com/prometheus/alertmanager/blob/c0a7b75c9cfb0772bdf5ec7362775f5f7798a3a0/notify/webhook/webhook.go#L64
type Message struct {
	Version         string `json:"version"`
	GroupKey        string `json:"groupKey"`
	TruncatedAlerts uint64 `json:"truncatedAlerts"`
	// the remaining fields map to
	// https://github.com/prometheus/alertmanager/blob/c0a7b75c9cfb0772bdf5ec7362775f5f7798a3a0/template/template.go#L236
	Receiver string  `json:"receiver"`
	Status   string  `json:"status"`
	Alerts   []Alert `json:"alerts"`

	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}

// Alert maps to https://github.com/prometheus/alertmanager/blob/c0a7b75c9cfb0772bdf5ec7362775f5f7798a3a0/template/template.go#L249
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint,omitempty"`
}

// MessageResponse is returned after a successful entry into the store from a received Message
type MessageResponse struct {
	ID string `json:"id"`
}

// MessageEntry is saved prior to the return of a MessageResponse
type MessageEntry struct {
	ID     string  `json:"id"`
	Alerts []Alert `json:"alerts"`
}

func (m Message) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%#v", m)
	}
	return string(b)
}
