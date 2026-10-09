# テスト（TESTING）

Pulse Monitor は **3 サービス** が別々のテストランナー・lint ツールで検証されます。ローカル検証と CI の対応関係・各サービス固有の勘所をここに集約します。

Makefile / CI ワークフロー / `CONTRIBUTING.md` に散らばっていた情報の統合インデックスとして参照してください。

## サービスとテストランナー

| サービス | 言語 | テストランナー | Lint | 代表コマンド（Makefile） |
|---|---|---|---|---|
| `api-gateway/` | Python / FastAPI | `pytest` | `flake8` | `make test-python` / `make lint-python` |
| `health-checker/` | Go | 標準 `go test` | `go vet` | `make test-go` / `make lint-go` |
| `alert-service/` | TypeScript / Node | `jest`（`npm test`） | `npm run lint`（ESLint） | `make test-ts` / `make lint-ts` |

### Lint 設定の集約先

- Python: `api-gateway/.flake8`（Makefile は `flake8 app/ tests/ --max-line-length=120` を実行）
- Go: 標準 `go vet`（追加設定なし）
- TypeScript: `alert-service/.eslintrc*`（`npm run lint` が `package.json` 側の設定を参照）

## ローカルで CI と等価な検証を行う

Makefile の集約ターゲットで CI の各ジョブを再現できます：

```sh
make test        # test-python + test-go + test-ts
make lint        # lint-python + lint-go + lint-ts
```

Docker ベースで全サービスを起動したうえで `/health` エンドポイントの応答を確認する smoke test は Makefile の `health` ターゲットで実行できます：

```sh
make up          # docker compose up -d --build
make health      # 各サービスの /health を叩いて JSON を整形表示
```

`make health` はユニット / 統合テストの代わりではなく、「3 サービスが同時に起動していて疎通している」という最後の確認に使うものです。CI では docker compose を使った疎通テストは走らないため、ローカル固有の確認手段として位置付けてください。

## 新しいテストを追加する時のチェックリスト

### Python (`api-gateway/`)

- [ ] テストは `api-gateway/tests/` 配下、ファイル名は `test_*.py`（`pytest` のデフォルト discovery に合わせる）
- [ ] FastAPI のルートは `TestClient` を使ってインプロセスで検証する（外部ネットワーク呼び出しをしない）
- [ ] 時刻依存・HTTP I/O はモック化し、CI での flakiness を避ける
- [ ] `flake8 --max-line-length=120` を通す（`make lint-python` と同条件）

### Go (`health-checker/`)

- [ ] テストファイル名は `<対象>_test.go`、関数は `TestXxx(t *testing.T)` の規約に従う
- [ ] 並行処理を伴うコードを追加した時は `go test -race ./...` を手元で一度は走らせる
- [ ] テーブル駆動テスト + `t.Run(name, ...)` のサブテスト名でケースを識別可能にする
- [ ] `go vet ./...` を通す（`make lint-go` と同条件）

### TypeScript (`alert-service/`)

- [ ] テストは `jest` のデフォルト設定に従って配置する（`__tests__/` or `*.test.ts`）
- [ ] モックは `jest.mock(...)` を使い、`beforeEach` で `jest.resetModules()` / `jest.clearAllMocks()` を呼んで副作用漏れを防ぐ
- [ ] 型エラーをテストで隠さない（`as any` でなく `satisfies` or 型拡張で解く）
- [ ] `npm run lint` を通す

### 共通

- [ ] CI 3 ジョブすべてが新規テストで緑であることを `make test` でローカル確認してから push する
- [ ] 外部サービス（Postgres / Redis / 外部 API）に依存する統合テストは `docker-compose.yml` と整合させ、必要なら `make up` でローカル環境を立ち上げてから走らせる
- [ ] テストクリーンアップが必要な場合は `make clean` で `__pycache__` / `node_modules` / `health-checker` バイナリをまとめて削除できる

## 関連ドキュメント

- [`../CONTRIBUTING.md`](../CONTRIBUTING.md) — ブランチ運用・コミット規則・レビューの流れ
- [`./ARCHITECTURE.md`](./ARCHITECTURE.md) — 3 サービスの責務とサービス間通信（テストの境界設計を考える時の参照先）
- [`./METRICS.md`](./METRICS.md) — 何を測るか（統合テストで検証したい健全域の定義）
- [`./ALERTING.md`](./ALERTING.md) — アラートルール（E2E テストで検証したい発火条件）
- [`./TROUBLESHOOTING.md`](./TROUBLESHOOTING.md) — テスト以外の運用で発生しがちな事象の切り分け
- [`../Makefile`](../Makefile) — 本ドキュメントが参照する全ターゲットの一次定義
- [`../.github/workflows/ci.yml`](../.github/workflows/ci.yml) — CI の一次定義
