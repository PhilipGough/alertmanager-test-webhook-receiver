PROJECT_PATH := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
IMAGE?=quay.io/philipgough/alertmanager-test-webhook-receiver
TAG ?=  $(shell git -C $(PROJECT_PATH) rev-parse HEAD)

.PHONY: build
build:
	go build -o webhook $(PROJECT_PATH)/cmd/server/main.go

.PHONY: image
image:
	docker build -t $(IMAGE):$(TAG) -f $(PROJECT_PATH)/Dockerfile .

.PHONY: test
test:
	go test $(PROJECT_PATH)/...
