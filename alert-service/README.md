# Alert Service

Pulse Monitor プラットフォームのアラート/通知サービス。アラートルールの管理と、ルールに基づく通知 (Webhook) の発火・履歴管理を担う。

プラットフォーム全体の概要・アーキテクチャ図・API 仕様の詳細はリポジトリルートの [`README.md`](../README.md) と [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) を参照してください。本 README はこのディレクトリに閉じた情報だけを扱います。

## 技術スタック

| 項目 | 内容 |
|------|------|
| 言語 | TypeScript |
| ランタイム | Node.js 20 (Docker イメージは `node:20-alpine`) |
| フレームワーク | Express 4 |
| テスト / Lint | Jest, supertest, ESLint (`@typescript-eslint`) |
| ポート | `8002` (環境変数 `ALERT_PORT` で変更可) |

依存バージョンの正本は [`package.json`](package.json) / `package-lock.json` です。

## ディレクトリ構成

```
alert-service/
├── Dockerfile            # 本番用イメージ (非 root ユーザ node で起動)
├── package.json          # 依存 / スクリプト / Jest 設定
├── tsconfig.json         # TypeScript コンパイラ設定
├── jest.config.js        # Jest 設定
├── .eslintrc.json        # ESLint 設定
└── src/
    ├── index.ts              # エントリポイント (サーバ起動 / graceful shutdown)
    ├── app.ts                # Express アプリ定義 / ルーティング
    ├── alerts.ts             # アラート / ルールのビジネスロジック
    ├── config.ts             # 環境変数の読み込み・バリデーション
    ├── logger.ts             # 構造化ロガー
    ├── security-headers.ts   # セキュリティレスポンスヘッダ ミドルウェア
    └── *.test.ts             # 対応するユニットテスト
```

## ローカル実行

### 1. ルートから docker compose で起動する (推奨)

```bash
# リポジトリルートで
make up           # 3 サービス一括起動
curl http://localhost:8002/health
```

### 2. このディレクトリ単体で Node から起動する

```bash
cd alert-service
npm install
npm run dev        # ts-node で src/index.ts を起動
# or
npm run build && npm start
```

### 3. このディレクトリの Dockerfile だけをビルドして起動する

```bash
cd alert-service
docker build -t pulse-alert-service .
docker run --rm -p 8002:8002 pulse-alert-service
```

## テスト / Lint

Makefile 経由 (リポジトリルートで実行):

```bash
make test-ts    # jest --forceExit
make lint-ts    # eslint src/ --ext .ts
```

直接実行:

```bash
cd alert-service
npm test
npm run lint
```

## 公開エンドポイント (要約)

| Method | Path | 概要 |
|--------|------|------|
| `GET` | `/health` | ヘルスチェック |
| `POST` | `/rules` | アラートルール作成 |
| `GET` | `/rules` | アラートルール一覧 |
| `GET` | `/rules/{id}` | アラートルール詳細 |
| `DELETE` | `/rules/{id}` | アラートルール削除 |
| `POST` | `/alerts` | アラート発火 |
| `GET` | `/alerts` | アラート一覧 |
| `GET` | `/alerts/{id}` | アラート詳細 |
| `PUT` | `/alerts/{id}/resolve` | アラート解決 |

リクエスト/レスポンス形式の詳細、重複抑制 (deduplication) の仕様、Webhook 送信の挙動などはルート [`README.md`](../README.md#alert-service-8002) と [`docs/ALERTING.md`](../docs/ALERTING.md) を参照してください。

## 主要な環境変数

| 変数 | 既定値 | 説明 |
|------|--------|------|
| `ALERT_PORT` | `8002` | 待ち受けポート |
| `LOG_LEVEL` | `INFO` | ログレベル (`DEBUG` / `INFO` / `WARN` / `ERROR`) |
| `API_GATEWAY_URL` | `http://api-gateway:8000` | API Gateway の URL |

全体の環境変数とデフォルト値はリポジトリルートの [`.env.example`](../.env.example) と [`README.md`](../README.md#environment-variables) が一次情報です。

## 関連ドキュメント

- プラットフォーム全体: [`../README.md`](../README.md)
- アーキテクチャ: [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md)
- アラート運用: [`../docs/ALERTING.md`](../docs/ALERTING.md)
- テスト運用: [`../docs/TESTING.md`](../docs/TESTING.md)
- 障害対応: [`../docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md)
- コントリビュート: [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
