# Health Checker

Pulse Monitor プラットフォームのヘルスチェック実行サービス。登録済みサービスへ HTTP ヘルスチェックを実施し、結果とレイテンシを返す。

プラットフォーム全体の概要・アーキテクチャ図・API 仕様の詳細はリポジトリルートの [`README.md`](../README.md) と [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) を参照してください。本 README はこのディレクトリに閉じた情報だけを扱います。

## 技術スタック

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.22 |
| ルータ | [`go-chi/chi`](https://github.com/go-chi/chi) v5 |
| ロガー | 標準ライブラリ `log/slog` (JSON ハンドラ) |
| テスト / Lint | `go test`, `go vet` |
| ポート | `8001` (環境変数 `CHECKER_PORT` で変更可) |

依存バージョンの正本は [`go.mod`](go.mod) / [`go.sum`](go.sum) です。

## ディレクトリ構成

```
health-checker/
├── Dockerfile            # マルチステージビルド (非 root ユーザで起動)
├── go.mod / go.sum       # Go モジュール定義
├── main.go               # エントリポイント / サーバ起動 / graceful shutdown
├── checker/              # ヘルスチェックのコアロジック (HTTP クライアント)
├── server/               # chi ルーティング / HTTP ハンドラ
└── config/               # 環境変数の読み込み・バリデーション
```

## ローカル実行

### 1. ルートから docker compose で起動する (推奨)

```bash
# リポジトリルートで
make up           # 3 サービス一括起動
curl http://localhost:8001/health
```

### 2. このディレクトリ単体で Go から起動する

```bash
cd health-checker
go run .
# 別シェルで
curl http://localhost:8001/health
```

### 3. このディレクトリの Dockerfile だけをビルドして起動する

```bash
cd health-checker
docker build -t pulse-health-checker .
docker run --rm -p 8001:8001 pulse-health-checker
```

## テスト / Lint

Makefile 経由 (リポジトリルートで実行):

```bash
make test-go    # go test ./... -v
make lint-go    # go vet ./...
```

直接実行:

```bash
cd health-checker
go test ./... -v
go vet ./...
```

## 公開エンドポイント (要約)

| Method | Path | 概要 |
|--------|------|------|
| `GET` | `/health` | ヘルスチェック |
| `POST` | `/check` | 任意の URL に対するオンデマンドヘルスチェック |

`/check` のリクエスト/レスポンス形式と `latency_ms` の単位 (ミリ秒) など詳細はルート [`README.md`](../README.md#health-checker-8001) を参照してください。

## 主要な環境変数

| 変数 | 既定値 | 説明 |
|------|--------|------|
| `CHECKER_PORT` | `8001` | 待ち受けポート |
| `LOG_LEVEL` | `INFO` | ログレベル (`DEBUG` / `INFO` / `WARN` / `ERROR`) |
| `CHECK_TIMEOUT` | `5s` | ヘルスチェック HTTP リクエストのタイムアウト |
| `API_GATEWAY_URL` | `http://api-gateway:8000` | API Gateway の URL |
| `ALERT_SERVICE_URL` | `http://alert-service:8002` | Alert Service の URL |

サーバ側の `READ_HEADER_TIMEOUT` / `READ_TIMEOUT` / `WRITE_TIMEOUT` / `IDLE_TIMEOUT` / `SHUTDOWN_TIMEOUT` も環境変数で調整可能です (詳細は [`config/`](config/) 参照)。全体の環境変数とデフォルト値はリポジトリルートの [`.env.example`](../.env.example) と [`README.md`](../README.md#environment-variables) が一次情報です。

## 関連ドキュメント

- プラットフォーム全体: [`../README.md`](../README.md)
- アーキテクチャ: [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md)
- テスト運用: [`../docs/TESTING.md`](../docs/TESTING.md)
- 障害対応: [`../docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md)
- コントリビュート: [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
