# トラブルシューティング

Pulse Monitor (`api-gateway` / `health-checker` / `alert-service`) の
ローカル開発 / Docker Compose 実行時によく遭遇する問題と対処方法を
まとめています。

- セットアップ: [../README.md](../README.md)
- 開発フロー: [../CONTRIBUTING.md](../CONTRIBUTING.md)
- セキュリティ問題の報告: [../SECURITY.md](../SECURITY.md)

## 1. `docker compose up` が失敗する / ポート競合

### 症状

```
Error response from daemon: ... port is already allocated
```

### 対処

`docker-compose.yml` の公開ポートが他プロセスに使われていないか確認:

```sh
lsof -i :8000    # api-gateway (Python/FastAPI)
lsof -i :8080    # health-checker (Go)
lsof -i :3001    # alert-service (Node.js/TypeScript)
```

該当プロセスを停止するか、`.env` (`.env.example` 参照) でポートを変更してください。
前回起動していたコンテナが残っている場合は `docker compose down` で完全停止してから再度起動します。

## 2. `api-gateway` (Python) のテストが失敗する

### 症状

- lint エラー (`ruff` / `flake8`)
- `pytest` が `ImportError` を返す

### 対処

`.tool-versions` 記載の Python 3.12 を使用してください。
CI と同じコマンドで再現できます:

```sh
cd api-gateway
pip install -r requirements.txt
pip install -r requirements-test.txt
pytest -v
```

## 3. `health-checker` (Go) のテストが失敗する

### 症状

- `go: cannot find main module`
- `go vet` が `undeclared name` を返す

### 対処

`health-checker` ディレクトリで実行してください:

```sh
cd health-checker
go mod download
go vet ./...
go test -v ./...
```

Go のバージョンは `.tool-versions` の **1.22 系** に揃えてください
(`go version` で確認)。

## 4. `alert-service` (TypeScript) のテストが失敗する

### 症状

- `npm ci` が lockfile mismatch で失敗
- `npm test` がタイムアウトする

### 対処

```sh
cd alert-service
rm -rf node_modules
npm ci
npm run lint
npm test
```

Node.js のバージョンは `.tool-versions` の **22 系** を使用してください
(`nvm use 22` 等で切り替え)。
`package-lock.json` を手動編集しないこと (Dependabot 経由か
`npm install <pkg>` で更新)。

## 5. サービス間の名前解決に失敗する

### 症状

コンテナのログに `connection refused` / `no such host` が出る。

### 原因と対処

コンテナ内では `localhost` / `127.0.0.1` ではなく、
`docker-compose.yml` で定義されたサービス名で参照する必要があります。

`.env.example` を参考に、以下のように設定してください:

```
API_GATEWAY_URL=http://api-gateway:8000
HEALTH_CHECKER_URL=http://health-checker:8080
ALERT_SERVICE_URL=http://alert-service:3001
```

ホスト OS から直接アクセスする場合のみ `localhost` を使用します。

## 6. Docker ビルドが極端に遅い / OOM

### 対処

- Docker Desktop / colima のメモリを 4 GB 以上に増やす
- `docker system prune -a` で不要なイメージ / キャッシュを削除
- `docker compose build --parallel` はメモリを多く消費するため、
  低スペック環境では逐次ビルドに切り替える

## 7. Alert 通知が届かない

### チェック項目

- `.env` の Webhook URL (Slack / PagerDuty 等) が正しく設定されているか
- `alert-service` のログにエラーが出ていないか (`docker compose logs alert-service`)
- `health-checker` から `alert-service` への疎通が取れているか
  (前節 "サービス間の名前解決" を参照)

## 8. CI が緑にならない

### チェックリスト

- Python (`api-gateway`): lint / `pytest -v` 全通過
- Go (`health-checker`): `go vet ./...` / `go test -v ./...` 両方通過
- TypeScript (`alert-service`): `npm run lint` / `npm test` 両方通過
- 依存ファイル (`requirements.txt` / `go.mod` / `package.json`)
  を更新した場合は対応するロックファイル / キャッシュキーも同時更新
- `.tool-versions` と CI (`.github/workflows/ci.yml`) の指定バージョンが
  一致しているか

CI 定義は [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)
にあります。

## 関連ドキュメント

- [../README.md](../README.md) — セットアップ手順
- [../CONTRIBUTING.md](../CONTRIBUTING.md) — 開発フロー
- [../SECURITY.md](../SECURITY.md) — セキュリティ問題の報告
- [../CHANGELOG.md](../CHANGELOG.md) — 変更履歴
