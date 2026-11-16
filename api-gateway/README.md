# API Gateway

Pulse Monitor プラットフォームの中央 API。監視対象サービスの登録・一覧・状態更新・削除を担う。

プラットフォーム全体の概要・アーキテクチャ図・API 仕様の詳細はリポジトリルートの [`README.md`](../README.md) と [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) を参照してください。本 README はこのディレクトリに閉じた情報だけを扱います。

## 技術スタック

| 項目 | 内容 |
|------|------|
| 言語 | Python 3.12 |
| フレームワーク | FastAPI (ASGI: uvicorn) |
| バリデーション | Pydantic v2 |
| HTTP クライアント | httpx |
| テスト / Lint | pytest, pytest-asyncio, flake8 |
| ポート | `8000` (環境変数 `API_PORT` で変更可) |

依存バージョンの正本は [`requirements.txt`](requirements.txt) です。

## ディレクトリ構成

```
api-gateway/
├── Dockerfile            # 本番用イメージ (非 root ユーザで起動)
├── requirements.txt      # ランタイム / テスト依存
├── app/
│   ├── main.py           # FastAPI エントリポイント / ルーティング / ミドルウェア
│   ├── config.py         # 環境変数の読み込み・バリデーション
│   ├── models.py         # Pydantic モデル (Service / AlertConfig など)
│   └── store.py          # インメモリストア
└── tests/                # pytest テスト
```

## ローカル実行

### 1. ルートから docker compose で起動する (推奨)

```bash
# リポジトリルートで
make up           # 3 サービス一括起動
curl http://localhost:8000/health
```

### 2. このディレクトリ単体で Python から起動する

```bash
cd api-gateway
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

### 3. このディレクトリの Dockerfile だけをビルドして起動する

```bash
cd api-gateway
docker build -t pulse-api-gateway .
docker run --rm -p 8000:8000 pulse-api-gateway
```

## テスト / Lint

Makefile 経由 (リポジトリルートで実行):

```bash
make test-python    # pytest -v
make lint-python    # flake8 app/ tests/ --max-line-length=120
```

直接実行:

```bash
cd api-gateway
pytest -v
flake8 app/ tests/ --max-line-length=120
```

## 公開エンドポイント (要約)

| Method | Path | 概要 |
|--------|------|------|
| `GET` | `/health` | ヘルスチェック |
| `POST` | `/services` | 監視対象サービスの登録 |
| `GET` | `/services` | 登録済みサービス一覧 |
| `GET` | `/services/{id}` | サービス詳細取得 |
| `PUT` | `/services/{id}/status` | ステータス更新 |
| `DELETE` | `/services/{id}` | サービス削除 |

リクエスト/レスポンス形式の詳細はルート [`README.md`](../README.md#api-gateway-8000) を参照してください。

## 主要な環境変数

| 変数 | 既定値 | 説明 |
|------|--------|------|
| `API_PORT` | `8000` | 待ち受けポート |
| `LOG_LEVEL` | `INFO` | ログレベル (`DEBUG` / `INFO` / `WARN` / `ERROR`) |
| `HEALTH_CHECKER_URL` | `http://health-checker:8001` | Health Checker サービスの URL |
| `ALERT_SERVICE_URL` | `http://alert-service:8002` | Alert Service の URL |

全体の環境変数とデフォルト値はリポジトリルートの [`.env.example`](../.env.example) と [`README.md`](../README.md#environment-variables) が一次情報です。

## 関連ドキュメント

- プラットフォーム全体: [`../README.md`](../README.md)
- アーキテクチャ: [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md)
- テスト運用: [`../docs/TESTING.md`](../docs/TESTING.md)
- 障害対応: [`../docs/TROUBLESHOOTING.md`](../docs/TROUBLESHOOTING.md)
- コントリビュート: [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
