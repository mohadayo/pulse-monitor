# よくある質問 (FAQ)

Pulse Monitor の設計判断・環境変数の意図・運用の疑問に答える Q&A 集です。
構造説明は [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) を、症状ベースの逆引きは [`docs/TROUBLESHOOTING.md`](./TROUBLESHOOTING.md) を参照してください。

## 1. アーキテクチャの設計判断

### Q1-1. なぜ 3 サービスに分けているのですか？

責務が独立しており、スケールと障害の隔離が別々に必要だからです。

- `api-gateway` — 外部 API 入口・認証・レート制御の境界
- `health-checker` — 外向きの監視ワーカ。高頻度の I/O を捌く
- `alert-service` — 通知の抑制・整形・配信の状態を持つ

これらを 1 プロセスに束ねると、監視ワーカのメモリ / ゴルーチン増加が API 応答性に波及したり、通知系のクラッシュが監視全体を止めることになります。

### Q1-2. なぜ Docker Compose 前提の構成ですか？

ローカル開発・ドキュメント化・ポート/ネットワーク定義の一元化が容易なためです。本番オーケストレーション基盤への移行時は、Compose の `docker-compose.yml` をベースに読み替えます。

### Q1-3. サービス間通信はどうやっていますか？

Compose のサービス名で解決される HTTP 呼び出しです。`.env.example` の `HEALTH_CHECKER_URL` / `ALERT_SERVICE_URL` / `API_GATEWAY_URL` がそれぞれの接続先です。

## 2. 言語選定に関する Q&A

### Q2-1. なぜ 3 つの言語を混在させているのですか？

各サービスの特性に最も適した言語を選択しているためです。

- **api-gateway (Python / FastAPI)** — 型注釈・OpenAPI 自動生成・エコシステムの成熟度が API 境界に適している
- **health-checker (Go)** — 高並列 I/O と低レイテンシが要求される監視ワーカに Go のゴルーチン / チャネルが自然
- **alert-service (TypeScript)** — 通知のテンプレートエンジンや Web 系配信 SDK が Node.js エコシステムに揃っている

### Q2-2. 統一した方が保守性が上がるのでは？

Yes / No 両面あります。この構成の狙いは「各サービスの `Dockerfile` と `README` 内で完結する」ことで、言語横断のレビュー負荷は上がるものの、サービス単位のオンボーディングは軽くなります。将来的にサービスを 2 言語程度に集約する選択肢は否定していません。

## 3. 環境変数の意図

### Q3-1. `SHUTDOWN_TIMEOUT=30s` はなぜ 30 秒？

Go の health-checker が「実行中のリクエストが完了するまで待つ」猶予です。ヘルスチェック 1 件あたり `READ_TIMEOUT=15s` + `WRITE_TIMEOUT=15s` を最大とみて、キューを空にする時間として 30 秒を選んでいます。

### Q3-2. `CHECK_TIMEOUT=5s` と他のタイムアウトの関係は？

- `CHECK_TIMEOUT` — health-checker が外部サービスへヘルスチェックを打つときの上限
- `READ_TIMEOUT` / `WRITE_TIMEOUT` — health-checker 自身が受け付ける HTTP の全体上限

つまり `CHECK_TIMEOUT` は「外向き I/O の上限」、`READ_TIMEOUT` / `WRITE_TIMEOUT` は「内向き HTTP サーバの上限」です。混同しないよう注意してください。

### Q3-3. `ALERT_DEDUP_WINDOW_SECONDS=300` を変更するときの注意は？

- **短くする (例: 60):** 通知の粒度が細かくなる代わりに、フラップ時に大量通知が飛ぶリスクが上がる
- **長くする (例: 900):** ノイズは減るが、新たな障害が発火するまでの気付きが遅れる
- **0 にする:** dedup を無効化する。テスト時のみ推奨

## 4. 運用の Q&A

### Q4-1. どのサービスから再起動すべきですか？

原則は影響が閉じたサービスから。監視の空白を作りたくなければ health-checker、通知系のみの復旧なら alert-service、外部影響がある場合は api-gateway 単独から。

### Q4-2. ヘルスチェック応答が遅い場合、まず何を疑うべき？

1. `CHECK_TIMEOUT` を超えている外部サービスが原因の場合が多い
2. 次に health-checker → api-gateway の疎通、DNS 解決、Compose ネットワーク
3. 最後に health-checker 自体の CPU / メモリ

### Q4-3. アラートが重複するときは？

`ALERT_DEDUP_WINDOW_SECONDS` を確認し、想定より短ければ延長を検討します。それでも重複する場合は Alert Service のログで `serviceId + message` のキーで dedup が効いているか確認してください。

## 5. その他

### Q5-1. どこから貢献を始めればよいですか？

[`CONTRIBUTING.md`](../CONTRIBUTING.md) を読んだ上で、`good first issue` ラベルが付いた Issue を優先的に見ていただければと思います。

### Q5-2. セキュリティ脆弱性はどこに報告すればよいですか？

[`SECURITY.md`](../SECURITY.md) の手順に従ってください。公開 Issue には投稿しないでください。

## 6. 関連ドキュメント

- [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) — システム全体像・データフロー
- [`docs/TROUBLESHOOTING.md`](./TROUBLESHOOTING.md) — 症状 → 原因 → 対処の逆引き
- [`README.md`](../README.md) — 概要 / セットアップ
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — コントリビュート手順
- [`SECURITY.md`](../SECURITY.md) — 脆弱性報告
