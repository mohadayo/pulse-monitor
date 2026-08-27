# アーキテクチャ設計

`pulse-monitor` は複数サービス (api-gateway / health-checker / alert-service) をヘルス監視するプラットフォーム。全体像・責務・データフロー・拡張ポイントをここに集約する。

- 個別の障害調査は `docs/TROUBLESHOOTING.md`

本書は "システム全体を短時間で理解する" ためのマップとし、実装詳細は各サービスの README と関連ドキュメントに委ねる。

## 目次

- [1. 全体構成](#1-全体構成)
- [2. サービス別 責務](#2-サービス別-責務)
- [3. データフロー](#3-データフロー)
- [4. 主要データモデル](#4-主要データモデル)
- [5. 拡張ポイント](#5-拡張ポイント)
- [6. 非機能要件](#6-非機能要件)

## 1. 全体構成

```
+-----------------+        +--------------------+        +------------------+
|  監視対象サービス | <----- |  health-checker    | -----> |  api-gateway     |
|  (HTTP/TCP/etc) |  Probe |  (Go)              |  結果  |  (Python)        |
+-----------------+        +--------------------+        +------------------+
                                                                 |
                                                          異常検知 |
                                                                 v
                                                        +------------------+
                                                        |  alert-service   |
                                                        |  (TypeScript)    |
                                                        +------------------+
                                                                 |
                                                          通知チャネル (Slack/Email/PagerDuty など)
```

言語選定の意図:

- **health-checker (Go)**: 大量のプローブを並行実行しても軽量。goroutine と静的バイナリで運用がシンプル
- **api-gateway (Python)**: 集計・時系列処理・ダッシュボード向け API に強く、pandas / FastAPI の生産性が高い
- **alert-service (TypeScript)**: 通知プロバイダの SDK が最も揃っており、テンプレートエンジン (JSX ライク) が扱いやすい

`docker-compose.yml` で 3 サービスをまとめて起動できる。

## 2. サービス別 責務

### 2.1 health-checker (Go)

- 監視対象のリストを設定から読み、周期的にプローブ (HTTP / TCP / ICMP) を打つ
- 結果を api-gateway へ push、または api-gateway に pull させる
- 直近の状態をメモリ上に保持し、`/metrics` で公開
- 単一責務: "プローブを打ち、結果を出す"。判定ロジックは持たない

### 2.2 api-gateway (Python)

- health-checker からの結果を受け取り、時系列で保持
- ダッシュボード向け REST API を提供 (現在状態・履歴・SLI 集計)
- 異常判定 (連続失敗回数、レイテンシ閾値、SLO 消費率) を実施し、alert-service にイベントを渡す
- 判定ロジックとダッシュボード用集計を集中させる

### 2.3 alert-service (TypeScript)

- api-gateway からの異常イベントを受け取り、深刻度・ルーティングに応じて通知
- 抑制 (deduplication)・スロットリング・再送ポリシーをここで管理
- 通知先チャネル (Slack / Email / PagerDuty など) の抽象を持ち、プロバイダを差し替えやすくする

## 3. データフロー

### 3.1 通常監視 (Steady state)

```
health-checker.probe()
  └─ push 結果 ─→ api-gateway.ingest()
                     └─ 時系列に保存 / /metrics 更新
                            └─ ダッシュボード API へ提供
```

### 3.2 異常検出 (Alert firing)

```
api-gateway.evaluate()
  └─ 閾値超過を検知
        └─ AlertEvent を作成 ─→ alert-service.dispatch()
                                    └─ 抑制 / スロットリング判定
                                          └─ 通知チャネルへ送出
```

### 3.3 回復通知 (Alert resolved)

```
api-gateway.evaluate()
  └─ 連続正常回数が閾値を超過
        └─ ResolvedEvent を作成 ─→ alert-service.dispatch()
                                       └─ 同一 alert_id に紐づく "解決" 通知を送出
```

## 4. 主要データモデル

### ProbeResult (health-checker → api-gateway)

| フィールド | 型 | 説明 |
| :-- | :-- | :-- |
| `target_id` | string | 監視対象の識別子 (URL や名称) |
| `probe_type` | enum | `http` / `tcp` / `icmp` |
| `ts` | RFC3339 | プローブ実行時刻 |
| `success` | bool | 成功したか |
| `latency_ms` | int | 応答時間 (成功時のみ) |
| `error` | string | エラーメッセージ (失敗時のみ) |

### AlertEvent (api-gateway → alert-service)

| フィールド | 型 | 説明 |
| :-- | :-- | :-- |
| `alert_id` | string | 一意の識別子 (target_id + rule の派生) |
| `severity` | enum | `critical` / `warning` / `info` |
| `state` | enum | `firing` / `resolved` |
| `summary` | string | 短い要約 (通知タイトル向け) |
| `annotations` | object | 詳細 (対応手順への導線など) |
| `fired_at` / `resolved_at` | RFC3339 | 発火・回復時刻 |

## 5. 拡張ポイント

### 5.1 監視対象の追加

- health-checker の設定に対象を追加する (プローブ種別・周期・タイムアウト)
- 追加のコード変更は不要 (設定駆動)

### 5.2 プローブ種別の追加

- health-checker 内で新しい probe 実装を追加する
- 出力は既存の `ProbeResult` に揃える。api-gateway 側の変更を不要にする

### 5.3 通知チャネルの追加

- alert-service にプロバイダアダプタを追加する
- ルーティング設定で "どの severity をどのチャネルに流すか" を宣言的に設定できるようにする

### 5.4 判定ルールの追加

- api-gateway に評価ロジックを追加する
- ルール宣言は設定 (YAML/JSON) から読み込み、コード変更なしで運用側が編集できることを目指す

## 6. 非機能要件

### 観測

- 全サービスは JSON 構造化ログを stdout に出す (必須フィールド: `ts` / `level` / `service` / `msg` / `request_id`)
- Prometheus 形式の `/metrics` を全サービスが公開
- 相関 ID (`X-Request-Id`) を境界で受け渡す

### 可用性

- health-checker はステートレスに近い設計とし、再起動で即復旧できる
- api-gateway の時系列ストアは PV に永続化する
- alert-service は通知の "at-least-once" を担保。冪等キー (`alert_id + state`) で重複送信を吸収

### セキュリティ

- API キー・接続情報は `.env` に集約 (`.env.example` に必須キーの型のみ示す)
- CI で依存の既知脆弱性を継続監視
- SECURITY.md の報告経路を尊重

## 変更履歴

- 2026-08: 初版作成。
