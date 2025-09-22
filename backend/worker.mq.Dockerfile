FROM golang:1.24.5-alpine3.22 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o app ./cmd/worker/mq

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/app .

ENTRYPOINT ["./app"]