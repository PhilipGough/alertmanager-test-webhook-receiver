FROM golang:1.17-alpine as builder

WORKDIR /app

ENV CGO_ENABLED=0
COPY . .
RUN apk add make && \
    make build

FROM scratch as app
LABEL service="alertmanager-webhook"
COPY --from=builder /app/webhook /

ENTRYPOINT ["/webhook"]