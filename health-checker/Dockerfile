# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o health-checker .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates

# CIS Docker Benchmark 4.1 準拠: 非 root ユーザを作成
RUN adduser -D -H -s /sbin/nologin app

WORKDIR /app
COPY --from=builder --chown=app:app /app/health-checker .

USER app

EXPOSE 8001
CMD ["./health-checker"]
