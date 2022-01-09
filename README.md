# alertmanager-test-webhook-receiver

alertmanager-test-webhook-receiver is a small webserver that receives webhook notifications from Alertmanager
and saves the alerts received in memory or on disk.

The server can be used to facilitate e2e testing and for verification of alerting configuration/pipelines.

### Building

* Running `make build` outputs a `webhook` binary which can be run locally.
* Running `make image` builds a container image.

### Testing

* Run `make test` to run the tests.

The following snippet will build and start the webserver locally and allow you to send a sample payload and test the result:

```bash
make build
./webhook
curl -X POST -H "Content-Type: application/json" -d @./cmd/server/testdata/request.json localhost:8080/webhook
curl localhost:8080//history/Test_webhook
```
