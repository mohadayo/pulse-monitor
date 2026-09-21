# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o health-checker .

FROM alpine:3.20

# ランタイム前処理を 1 レイヤに統合する (hadolint DL3059):
#   - `ca-certificates` を追加 (TLS 検証用)
#   - CIS Docker Benchmark 4.1 準拠の非 root ユーザを作成
# apk バージョンの pin (DL3018) は `.hadolint.yaml` で ignore する
# (ca-certificates は最新を入れる方針)。
RUN apk --no-cache add ca-certificates \
    && adduser -D -H -s /sbin/nologin app

WORKDIR /app
COPY --from=builder --chown=app:app /app/health-checker .

USER app

EXPOSE 8001
CMD ["./health-checker"]
