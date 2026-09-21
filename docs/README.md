# Pulse Monitor ドキュメント目次

`docs/` 配下のドキュメントを目的別に俯瞰するためのインデックスです。ルート [`../README.md`](../README.md) は「Pulse Monitor とは何か / どう動かすか」の入り口、こちらは「何か知りたい・調べたい」ときの導線として機能します。

## 目的別ガイド

| やりたいこと | 参照先 |
| --- | --- |
| システム全体像・サービス責務・データフローを把握したい | [`ARCHITECTURE.md`](ARCHITECTURE.md) |
| アラートの設計方針・ルール管理・通知チャンネル・重複抑制を知りたい | [`ALERTING.md`](ALERTING.md) |
| 設定・運用・仕様に関する「よくある質問」を確認したい | [`FAQ.md`](FAQ.md) |
| 症状から障害切り分け手順を辿りたい・エラー時の対処を調べたい | [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) |

## 初めての方向け

初めて Pulse Monitor に触れる場合は、以下の順序で読むことを推奨します。

1. ルート [`../README.md`](../README.md) — プロジェクト概要・クイックスタート・API 概要
2. [`ARCHITECTURE.md`](ARCHITECTURE.md) — API Gateway (Python) / Health Checker (Go) / Alert Service (TypeScript) の 3 サービス構成を把握
3. [`ALERTING.md`](ALERTING.md) — 実際にアラートを設定・受信するときの考え方
4. [`FAQ.md`](FAQ.md) / [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) — 詰まったとき・想定挙動を確認したいときのリファレンス

## リポジトリ全体のガイド

`docs/` 以外にも、以下のリポジトリルート直下のドキュメントが対応するテーマを扱っています。ここでリンクしておくことで `docs/` からもたどれるようにします。

| テーマ | 参照先 |
| --- | --- |
| コントリビュート方針・PR / Issue 作成手順・開発フロー | [`../CONTRIBUTING.md`](../CONTRIBUTING.md) |
| コミュニティ規範（Contributor Covenant 準拠） | [`../CODE_OF_CONDUCT.md`](../CODE_OF_CONDUCT.md) |
| 脆弱性報告経路・サポート対象バージョン | [`../SECURITY.md`](../SECURITY.md) |
| 変更履歴（Keep a Changelog 形式） | [`../CHANGELOG.md`](../CHANGELOG.md) |
| 開発用 Make タスク一覧 | [`../Makefile`](../Makefile)（`make help` でも参照可） |
| CI ワークフロー本体 | [`../.github/workflows/ci.yml`](../.github/workflows/ci.yml) |

## 新しいドキュメントを追加する場合

- **プロジェクト全体に関するもの**（コントリビュート方針、セキュリティ方針、変更履歴、リリース手順など）はリポジトリルート直下に置きます。
- **開発・運用・障害対応のリファレンス**（アーキテクチャ図、機能設計、トラブルシューティング手順など）は `docs/` 配下に置きます。
- 新しいファイルを `docs/` に追加した場合は、本ファイルの「目的別ガイド」表にエントリを追加し、目的（読み手が「何を知りたい」ときに参照するのか）を 1 行で書き添えてください。
- 命名は原則 `UPPER_SNAKE_CASE.md`（既存 `ARCHITECTURE.md` / `ALERTING.md` などに合わせる）とします。
- ドキュメント間の関係は、必要に応じて「関連ドキュメント」節で相互リンクしてください。
