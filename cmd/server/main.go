package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"text/template"
	"time"

	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/api"
	"github.com/philipgough/alertmanager-test-webhook-receiver/pkg/store"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/gorilla/mux"
)

var (
	listenAddress string
	logLevel      string
	storeIDTmpl   string
	dbPath        string
)

const (
	defaultListenAddress   = ":8080"
	defaultLogLevel        = "info"
	defaultStoreIDTemplate = `{{ .GroupLabels.alertname }}_{{ .Receiver }}`
	defaultDbPath          = ""
)

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&listenAddress, "listen.address", defaultListenAddress, "The network address to listen on")
	flagset.StringVar(&logLevel, "log.level", defaultLogLevel, "One of 'debug', 'info', 'warn', 'error'")
	flagset.StringVar(&storeIDTmpl, "id.template", defaultStoreIDTemplate, "The template used to generate the ID for storage")
	flagset.StringVar(&dbPath, "db.path", defaultDbPath, "The file path to the history store. Empty (default) uses in-memory store")

	flagset.Parse(os.Args[1:])

	logger := setupLogger(logLevel)
	store, err := store.NewKeyValueStore(dbPath, logger)
	if err != nil {
		level.Error(logger).Log("msg", "failed to initialise database", "err", err)
		os.Exit(1)
	}

	srv := &server{
		logger:      logger,
		store:       store,
		router:      mux.NewRouter(),
		idGenerator: buildIdGenerator(storeIDTmpl),
	}

	go func() {
		if err := srv.run(listenAddress); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				level.Error(logger).Log("msg", "server run returned an error", "err", err)
				os.Exit(1)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	level.Info(logger).Log("msg", "interrupt received. shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	srv.close(ctx)

	level.Info(logger).Log("msg", "exiting...")
	os.Exit(0)
}

type server struct {
	logger log.Logger
	store  store.Store
	router *mux.Router
	srv    *http.Server
	idGenerator
}

func (s *server) run(address string) error {
	s.srv = &http.Server{Addr: address, Handler: s.router}
	s.routes()
	level.Info(s.logger).Log("msg", "server starting", "address", address)
	return s.srv.ListenAndServe()
}

func (s *server) routes() {
	s.router.HandleFunc("/webhook", s.handleWebhook()).Methods(http.MethodPost)
	s.router.HandleFunc("/history/{id}", s.handleHistory()).Methods(http.MethodGet)
	s.router.HandleFunc("/history", s.handleListHistory()).Methods(http.MethodGet)
}

func (s *server) close(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *server) handleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var into api.Message
		if err := json.NewDecoder(r.Body).Decode(&into); err != nil {
			level.Error(s.logger).Log("msg", "failed to decode JSON body", "err", err)
			http.Error(w, "failed to decode JSON body", http.StatusBadRequest)
			return
		}

		level.Debug(s.logger).Log("msg", "webhook received", "data", into)

		id, err := s.idGenerator(into)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to generate ID for store", "err", err)
			http.Error(w, "failed to generate ID from request body", http.StatusInternalServerError)
			return
		}

		if err := s.store.Set(id, into.Alerts); err != nil {
			level.Error(s.logger).Log("msg", "failed to save alerts", "id", id, "err", err)
			http.Error(w, "failed to save webhook info", http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(api.MessageResponse{ID: id})
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to encode response", "id", id, "err", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

func (s *server) handleHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := mux.Vars(r)["id"]
		history, err := s.store.Get(id)
		if err != nil {
			status := http.StatusInternalServerError
			level.Error(s.logger).Log("msg", "failed to read webhook history", "id", id, "err", err)
			if errors.Is(err, store.ErrNotFound) {
				status = http.StatusNotFound
			}
			http.Error(w, "failed to read webhook history", status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(history); err != nil {
			level.Error(s.logger).Log("msg", "failed to encode webhook history", "id", id, "err", err)
			http.Error(w, "failed to encode webhook history", http.StatusInternalServerError)
			return
		}

		return
	}
}

func (s *server) handleListHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		history, err := s.store.List()
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to list webhook history", "err", err)
			http.Error(w, "failed to list webhook history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(history); err != nil {
			level.Error(s.logger).Log("msg", "failed to encode the list webhook history", "err", err)
			http.Error(w, "failed to encode the list webhook history", http.StatusInternalServerError)
			return
		}

		return
	}
}

type idGenerator func(payload api.Message) (string, error)

func buildIdGenerator(tmpl string) idGenerator {
	t := template.Must(template.New("html-tmpl").Parse(tmpl))
	return func(payload api.Message) (string, error) {
		w := bytes.NewBuffer([]byte{})
		if err := t.Execute(w, payload); err != nil {
			return "", err
		}
		return w.String(), nil
	}
}

func setupLogger(lvl string) log.Logger {
	var (
		logger    log.Logger
		lvlOption level.Option
	)

	switch strings.ToLower(lvl) {
	case "debug":
		lvlOption = level.AllowDebug()
	case "info":
		lvlOption = level.AllowInfo()
	case "warn":
		lvlOption = level.AllowWarn()
	case "error":
		lvlOption = level.AllowError()
	}

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = level.NewFilter(logger, lvlOption)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	return logger
}
