# メトリクス / SLI リファレンス

Pulse Monitor が扱う主要メトリクスの意味・取得元・観測方法と、推奨されるサービスレベル指標 (SLI) / 目標値 (SLO) の目安をまとめたリファレンスです。

- **[`ALERTING.md`](ALERTING.md)** は「どのメトリクスをどう組み合わせてアラート化するか」の運用面を扱います。
- **[`TROUBLESHOOTING.md`](TROUBLESHOOTING.md)** は「症状から障害を切り分ける」対処面を扱います。
- 本ドキュメントはその上流にあたる「Pulse Monitor は何を測っており、どの水準までが健全か」を定義することを目的とします。

具体的なアラート閾値・通知チャンネル設計は `ALERTING.md`、障害時の対処手順は `TROUBLESHOOTING.md` にそれぞれ委譲します。

## 主要メトリクス一覧

以下は Pulse Monitor の 3 サービス（API Gateway / Health Checker / Alert Service）が公開・記録するメトリクスの俯瞰です。

| # | メトリクス | 出力元 | 型 | 単位 | 取得エンドポイント / 参照 |
|---|---|---|---|---|---|
| 1 | サービス `status` | API Gateway / Health Checker | 列挙型 | `healthy` / `unhealthy` / `unknown` | `GET /services`、`POST /check` |
| 2 | HTTP `status_code` | Health Checker | 整数 | HTTP ステータスコード | `POST /check` |
| 3 | `latency_ms` (ラウンドトリップ) | Health Checker | 整数 | ミリ秒 | `POST /check` |
| 4 | `interval_seconds` (チェック間隔) | API Gateway | 整数 | 秒 | `POST /services`、`GET /services/{id}` |
| 5 | `last_checked` | API Gateway | ISO 8601 | 日時 | `GET /services/{id}` |
| 6 | 登録サービス数 | API Gateway | 整数 | 件 | `GET /services` の要素数 |
| 7 | アラートルール数 | Alert Service | 整数 | 件 | `GET /rules` の要素数 |
| 8 | 発火中アラート数 | Alert Service | 整数 | 件 | `GET /alerts` の要素数（未解決のもの） |
| 9 | サービス自身の `/health` 応答 | 全サービス | 論理値 / HTTP 200 | — | 各サービスの `GET /health` |

## メトリクス詳細

### 1. サービス `status`

- **意味**: Pulse Monitor が観測している対象サービスの「直近チェック時点」の健全性ラベル。
- **値**:
  - `healthy` — 直近のヘルスチェックが成功（HTTP 2xx かつタイムアウトなし）。
  - `unhealthy` — 直近のヘルスチェックが失敗（非 2xx、タイムアウト、接続不能など）。
  - `unknown` — 未チェック、または結果が確定していない（登録直後の初期値）。
- **取得**: `GET /services` の各要素の `status` フィールド。
- **注意**: `status` は「直近 1 回の結果」を反映するのみで、時系列のばたつき（フラッピング）は表現しません。フラッピング検知が必要な場合は連続失敗回数を別途集計してください（詳細は `ALERTING.md` 参照）。

### 2. HTTP `status_code`

- **意味**: Health Checker が対象 URL へリクエストした際に受信した HTTP ステータスコード。
- **正常判定**: 2xx を `healthy`、それ以外を `unhealthy` と扱うのが既定の運用です。
- **典型的な観測パターン**:
  - `200` — 完全正常。
  - `301` / `302` — リダイレクト。監視 URL は原則リダイレクト先を直接指すよう構成してください。
  - `401` / `403` — 認可設定漏れ、または監視用エンドポイントが認証必須になっている。
  - `5xx` — 対象サービス側の障害。`TROUBLESHOOTING.md` を参照。
  - `0` またはフィールド欠落 — 接続失敗 / タイムアウト（後述の `latency_ms` と合わせて判断）。

### 3. `latency_ms`（ラウンドトリップ時間）

- **意味**: Health Checker から対象 URL への HTTP リクエストのラウンドトリップ時間（ミリ秒）。
- **注意**: これは Pulse Monitor から見た **クライアント側計測** であり、対象サービス内部の処理時間ではありません。Pulse Monitor と対象サービスの間のネットワーク経路の遅延も含みます。
- **推奨レンジ（同一リージョン内、外部 URL の場合）**:
  - `< 200 ms` — 良好。
  - `200 – 1000 ms` — 要観測。継続的に高い場合はネットワーク経路または対象サービスの負荷を確認。
  - `> 1000 ms` — 悪化。`CHECK_TIMEOUT`（既定 `5s`）に近づいており、間欠的な `unhealthy` が発生しやすい水準です。

### 4. `interval_seconds`（チェック間隔）

- **意味**: 対象サービスに対して Health Checker がヘルスチェックを実行する周期（秒）。
- **設計上の注意**:
  - 短くしすぎると対象サービスと Pulse Monitor 双方の負荷が線形に増加します。
  - 長くしすぎると `MTTD`（障害検知までの平均時間）が伸びます。
  - **MTTD の下限は `interval_seconds`** であり、加えて連続失敗判定を挟む場合はその分だけ検知が遅延します。

### 5. `last_checked`

- **意味**: 対象サービスに対して最後にヘルスチェックが実行された時刻 (ISO 8601, UTC)。
- **観測上の使い方**: `now() - last_checked` が `interval_seconds` を大きく上回っている場合、Health Checker 側のスケジューラ停滞を疑ってください。

### 6 – 8. 登録数系メトリクス

| メトリクス | 想定される観測意義 |
|---|---|
| 登録サービス数 (`GET /services` 要素数) | Pulse Monitor の管理対象規模。急激な減少は登録誤削除、急増は自動登録の暴走を示唆。 |
| アラートルール数 (`GET /rules` 要素数) | ルール未設定のサービスがある / ルール多重登録の傾向を検出。 |
| 発火中アラート数 (未解決 `GET /alerts` 要素数) | 現在の障害圧力の総量。過去分の未クローズが積み上がる場合は運用フロー側の見直しが必要。 |

### 9. サービス自身の `/health`

Pulse Monitor の 3 サービスは各自 `GET /health` を公開しており、`docker compose` 上でも `make health` から一括確認できます。監視対象を監視する構成上、Pulse Monitor 自身の `/health` が最も上位の SLI となります（後述）。

## SLI / SLO 目安

以下はプロジェクトの既定運用における推奨値です。プロダクション固有の要件に合わせて上書きしてください。

### Pulse Monitor 自身に対する SLI（メタ監視）

| SLI | 定義 | 推奨 SLO |
|---|---|---|
| **可用性** | 各サービス `/health` が HTTP 200 を返した割合 | 30 日ローリングで `>= 99.5%` |
| **チェック実行遅延** | `now() - last_checked` の 95 パーセンタイル | `<= 2 * interval_seconds` |
| **アラート配送遅延** | `unhealthy` 判定から Alert Service `POST /alerts` 受信までの中央値 | `<= 30 秒` |

### 監視対象サービスに対する SLI（Pulse Monitor が観測するもの）

| SLI | 定義 | 目安（既定） |
|---|---|---|
| **成功率** | `status == healthy` の割合 | `>= 99.9%` を SLO とする例が一般的 |
| **応答時間** | `latency_ms` の 95 パーセンタイル | サービス性質に依存。例: Web API なら `<= 500 ms` |
| **エラー予算** | (1 − SLO) × 期間 | 30 日 × `0.1%` ≒ 約 43 分 |

SLO を割り込む可能性が高まったとき（エラー予算の消費が急である、95p レイテンシが SLO 目安を超える等）にアラートを設定する方針は、`ALERTING.md` の「アラート設計指針」節で扱います。

## メトリクスとアラートの関係

アラートは「メトリクスがある閾値を跨いだ」ことを検知する仕組みです。したがって:

1. **測定 (このドキュメント)** — 何を、どこから、どの単位で取得しているかを定義する。
2. **健全域の定義 (このドキュメント)** — 上記メトリクスの推奨レンジ・SLO を定義する。
3. **アラート化 ([`ALERTING.md`](ALERTING.md))** — 健全域を逸脱したときの通知ルール・重複抑制・宛先を定義する。
4. **切り分け ([`TROUBLESHOOTING.md`](TROUBLESHOOTING.md))** — アラートが鳴った / 症状が出たときの調査手順を定義する。

新しいメトリクスを追加した場合は、上記フロー全体に沿って各ドキュメントを更新してください。

## 関連ドキュメント

- [`ARCHITECTURE.md`](ARCHITECTURE.md) — 各メトリクスがどのサービス境界で生成されるかの俯瞰。
- [`ALERTING.md`](ALERTING.md) — 本ドキュメントで定義したメトリクスをどうアラート化するか。
- [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) — 各メトリクスが異常値を示したときの切り分け手順。
- [`FAQ.md`](FAQ.md) — メトリクス / 監視まわりのよくある質問。
- ルート [`../README.md`](../README.md) — API エンドポイントとレスポンス例。
