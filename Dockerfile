FROM golang:1.17.6-alpine3.15 as builder

WORKDIR /app

ENV CGO_ENABLED=0
COPY . .
RUN apk add make && \
    make build

FROM gcr.io/distroless/base-debian10 as app
LABEL service="alertmanager-webhook"
COPY --from=builder /app/webhook /

ENTRYPOINT ["/webhook"]