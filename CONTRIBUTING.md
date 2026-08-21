# コントリビューションガイド

Pulse Monitor へのコントリビュートに関心を持っていただきありがとうございます。
本ドキュメントは、Issue 起票・PR 作成・レビューまでの流れをまとめた開発者向けガイドです。

はじめての方は [`README.md`](README.md) と [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) にも
必ず目を通してください。セキュリティ上の脆弱性報告は [`SECURITY.md`](SECURITY.md) の手順に従ってください。

## 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [リポジトリ構成](#リポジトリ構成)
- [開発フロー](#開発フロー)
- [ブランチ命名規則](#ブランチ命名規則)
- [コミットメッセージ規約](#コミットメッセージ規約)
- [テストと静的解析](#テストと静的解析)
- [プルリクエスト](#プルリクエスト)
- [Issue 起票](#issue-起票)
- [行動規範](#行動規範)

## 開発環境のセットアップ

必要なツールチェインは [`.tool-versions`](.tool-versions) に固定されています。
[asdf](https://asdf-vm.com/) または [mise](https://mise.jdx.dev/) を利用すると一括で導入できます。

| ツール  | バージョン |
| ------- | ---------- |
| Python  | 3.12       |
| Go      | 1.22       |
| Node.js | 22         |

### 1. Docker Compose を利用する場合（推奨）

3 サービスをまとめて起動できます。

```bash
git clone https://github.com/mohadayo/pulse-monitor.git
cd pulse-monitor
cp .env.example .env
make up      # 全サービスをビルド & 起動
make health  # 各サービスの /health を叩いて疎通確認
make logs    # ログ追跡
make down    # 停止
```

### 2. ローカルで直接起動する場合

サービスごとに個別に起動できます。詳細は各ディレクトリの `README` または `Makefile` を参照してください。

```bash
# API Gateway (Python / FastAPI, :8000)
cd api-gateway
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000

# Health Checker (Go / Chi, :8001)
cd health-checker
go run ./...

# Alert Service (TypeScript / Express, :8002)
cd alert-service
npm install
npm run dev
```

## リポジトリ構成

```
pulse-monitor/
├── api-gateway/      # Python 3.12 / FastAPI — 中央 API
├── health-checker/   # Go 1.22 / Chi        — ヘルスチェック実行
├── alert-service/    # TypeScript / Express — アラート配信
├── docker-compose.yml
├── Makefile          # up / down / test / lint など
├── .env.example      # 環境変数サンプル
└── .github/          # Issue / PR テンプレート, CODEOWNERS, Dependabot
```

変更対象のサービスに応じて、対応する言語ツールチェインが必要になります。
横断的な変更（`docker-compose.yml` など）を行う場合は 3 サービス全てのテストを通してください。

## 開発フロー

1. 事前に [Issues](https://github.com/mohadayo/pulse-monitor/issues) を検索し、既存の議論・重複がないか確認する
2. 大きめの変更を行う前に Issue を起票して合意形成する（typo・軽微な修正はスキップ可）
3. `main` から作業ブランチを切る（[ブランチ命名規則](#ブランチ命名規則)）
4. 変更を加え、ローカルで `make test` と `make lint` をパスさせる
5. [コミットメッセージ規約](#コミットメッセージ規約) に沿ってコミットする
6. Fork または feature ブランチを push し、[PR テンプレート](.github/PULL_REQUEST_TEMPLATE.md) に従って PR を作成する
7. レビュー指摘に対応し、CI をグリーンにしてマージを待つ

## ブランチ命名規則

```
<type>/<short-description>
```

| プレフィックス | 用途                                     |
| -------------- | ---------------------------------------- |
| `feat/`        | 新機能追加                               |
| `fix/`         | バグ修正                                 |
| `docs/`        | ドキュメントのみの変更                   |
| `chore/`       | ビルド・CI・依存関係更新など雑務        |
| `refactor/`    | 挙動を変えないリファクタリング           |
| `test/`        | テストの追加・修正                       |

例: `feat/alert-webhook-retry`, `fix/health-checker-timeout`, `docs/contributing-guide`

## コミットメッセージ規約

[Conventional Commits](https://www.conventionalcommits.org/ja/v1.0.0/) に準拠します。

```
<type>(<scope>): <subject>

<body>

<footer>
```

- **type**: `feat` / `fix` / `docs` / `chore` / `refactor` / `test` / `ci` / `build` / `perf`
- **scope**: 影響を受けるサービス（`api-gateway`, `health-checker`, `alert-service`）または
  横断的な領域（`docker`, `ci`, `deps` など）。省略可
- **subject**: 変更内容を命令形で簡潔に（末尾ピリオドなし）
- **footer**: `Closes #123` のように関連 Issue を記載

例:

```
feat(alert-service): Slack Webhook のリトライを 3 回に増やす
fix(health-checker): タイムアウト時の goroutine リークを解消
docs: コントリビューションガイドを追加
```

## テストと静的解析

PR を出す前に、少なくとも以下がローカルでパスしていることを確認してください。
CI は同等のチェックを再実行します。

```bash
make test   # 3 サービスすべての単体テスト
make lint   # flake8 / go vet / eslint
```

個別に実行することもできます。

```bash
make test-python   # api-gateway: pytest
make test-go       # health-checker: go test ./...
make test-ts       # alert-service: npm test

make lint-python   # flake8 app/ tests/ --max-line-length=120
make lint-go       # go vet ./...
make lint-ts       # npm run lint
```

環境変数を追加・変更した場合は **必ず [`.env.example`](.env.example) を更新** してください。

## プルリクエスト

- PR タイトルはコミットメッセージ規約と同じ形式を推奨します（例: `docs: コントリビューションガイドを追加`）
- 本文は [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) が自動で挿入されるので、
  すべてのセクション（変更概要 / 対応 Issue / 影響範囲 / 動作確認手順 / チェックリスト）を埋めてください
- 破壊的変更を含む場合は、PR 本文と [`CHANGELOG.md`](CHANGELOG.md) の両方に明記してください
- レビュアーは [`.github/CODEOWNERS`](.github/CODEOWNERS) に基づき自動で割り当てられます
- レビュー指摘への修正は追加コミットで行い、マージ時に squash されます（履歴は PR ページで追えます）

## Issue 起票

- バグ報告: [`.github/ISSUE_TEMPLATE/bug_report.md`](.github/ISSUE_TEMPLATE/bug_report.md)
- 機能要望: [`.github/ISSUE_TEMPLATE/feature_request.md`](.github/ISSUE_TEMPLATE/feature_request.md)
- サポート／使い方の質問: [`.github/SUPPORT.md`](.github/SUPPORT.md) を参照
- セキュリティ脆弱性: 公開 Issue ではなく [`SECURITY.md`](SECURITY.md) の連絡先へ

## 行動規範

本プロジェクトのすべての活動には [Contributor Covenant v2.1](CODE_OF_CONDUCT.md) が適用されます。
参加者は互いに敬意を持って接することが求められます。

---

ご質問や不明点は Issue または Discussion にてお気軽にお寄せください。あなたのコントリビュートをお待ちしています。
