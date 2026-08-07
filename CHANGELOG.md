# Changelog

このプロジェクトの主な変更点を記録するファイルです。

フォーマットは [Keep a Changelog v1.1.0](https://keepachangelog.com/ja/1.1.0/) に、
バージョン番号は [Semantic Versioning](https://semver.org/lang/ja/) に準拠します。

## [Unreleased]

### Added

- （次回リリースで追加する機能をここに記載）

### Changed

- （挙動の変更をここに記載）

### Deprecated

- （非推奨になった機能をここに記載）

### Removed

- （削除された機能をここに記載）

### Fixed

- （バグ修正をここに記載）

### Security

- （セキュリティ関連の修正をここに記載）

## [0.1.0] - 2026-05-08

初回リリース。Pulse Monitor の Baseline 実装
（Python API Gateway / Go Health Checker / TypeScript Alert Service の 3 サービス構成）を記録します。

### Added

- **api-gateway (Python / FastAPI)**: 各サービスへの統合エントリポイント。
  監視対象・アラート情報を集約してクライアントに公開する。
- **health-checker (Go)**: 監視対象エンドポイントへの死活監視を
  非同期に実行し、結果をパーシストする。
- **alert-service (TypeScript / Node.js)**: 検知したインシデントを
  各種通知チャネル（Slack / Webhook 等）へ配信する。
- ローカル開発用の `docker-compose.yml` による 3 サービスの一括起動。
- 共通タスクを集約する `Makefile`。
- リポジトリ運用ドキュメント: `README.md` / `CODE_OF_CONDUCT.md`。
- 開発補助ファイル: `.gitignore` / `.env.example` / `.tool-versions`。
- **CI ワークフロー** (`.github/workflows/`):
  - Python (api-gateway) の lint (flake8) / test (pytest)
  - Go (health-checker) の build / test
  - TypeScript (alert-service) の lint (eslint) / test (jest)
- **Dependabot** による依存パッケージ (pip / npm / Go modules /
  GitHub Actions / Docker) の自動更新。

[Unreleased]: https://github.com/mohadayo/pulse-monitor/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mohadayo/pulse-monitor/releases/tag/v0.1.0
