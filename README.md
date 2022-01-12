# alertmanager-test-webhook-receiver

alertmanager-test-webhook-receiver is a small webserver that
[receives webhook notifications from Alertmanager](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)
and saves the alerts received in-memory or on disk.

The server can be used to facilitate e2e testing and for verification of alerting configuration/pipelines.

## Usage

The server exposes two endpoints:

[AlertManager](https://prometheus.io/docs/alerting/latest/alertmanager/) will send POST requests to the `/webhook` endpoint.

On receiving the webhook, the server will generate an ID from a [template](https://pkg.go.dev/text/template),
parsing the request body as defined [here](https://pkg.go.dev/github.com/prometheus/alertmanager@v0.23.0/notify/webhook#Message).
The generated ID will be used to store the list of [alerts](https://pkg.go.dev/github.com/prometheus/alertmanager@v0.23.0/template#Alert).

Stored data can be retrieved in one of the following ways:

An HTTP GET request to `/history/{id}` will return a list of alerts for that ID.
An HTTP GET request to `/history` will return a list of existing (ID, alerts) pairs.

### Configuration 
```shell
  -db.path string
        The file path to the history store. Empty (default) uses in-memory store
  -id.template string
        The template used to generate the ID for storage (default "{{ .GroupLabels.alertname }}_{{ .Receiver }}")
  -listen.address string
        The network address to listen on (default ":8080")
  -log.level string
        One of 'debug', 'info', 'warn', 'error' (default "info")
```

## Building

* Running `make build` outputs a `webhook` binary which can be run locally.
* Running `make image` builds a container image.

## Testing

* Run `make test` to run the tests.

### Smoke test

The following snippet will build and start the webserver locally and allow you to send a sample payload and test the result:

```bash
## Start the server with provided flags
make build && ./webhook --log.level=debug -id.template='{{ .GroupLabels.job }}_{{ .Receiver }}' --db.path=/tmp/history
## Execute a request using test payload to simulate webhook
curl -X POST -H "Content-Type: application/json" \
     -d @./cmd/server/testdata/request.json localhost:8080/webhook
## Read data from storage
curl -H "Content-Type: application/json" localhost:8080/history/prometheus24_webhook
```
