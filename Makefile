.DEFAULT_GOAL := help

.PHONY: help up down build test test-python test-go test-ts lint lint-python lint-go lint-ts logs health clean

help: ## このヘルプを表示する
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## 全サービスをビルドしてバックグラウンドで起動する (docker compose up --build -d)
	docker compose up --build -d

down: ## 全サービスを停止する (docker compose down)
	docker compose down

build: ## 全サービスのイメージをビルドする (docker compose build)
	docker compose build

test: test-python test-go test-ts ## 全サービスのテストを順に実行する (python / go / typescript)

test-python: ## API Gateway (Python) のテストを実行する (pytest)
	cd api-gateway && pip install -r requirements.txt -q && pytest -v

test-go: ## Health Checker (Go) のテストを実行する (go test ./...)
	cd health-checker && go test ./... -v

test-ts: ## Alert Service (TypeScript) のテストを実行する (jest)
	cd alert-service && npm install --silent && npm test

lint: lint-python lint-go lint-ts ## 全サービスの lint を順に実行する (python / go / typescript)

lint-python: ## API Gateway を flake8 で lint する
	cd api-gateway && flake8 app/ tests/ --max-line-length=120

lint-go: ## Health Checker を go vet で lint する
	cd health-checker && go vet ./...

lint-ts: ## Alert Service を eslint で lint する
	cd alert-service && npm run lint

logs: ## 全サービスのログを追従表示する (docker compose logs -f)
	docker compose logs -f

health: ## 3 サービスの /health エンドポイントを一括確認する
	@echo "API Gateway:" && curl -s http://localhost:8000/health | python3 -m json.tool
	@echo "Health Checker:" && curl -s http://localhost:8001/health | python3 -m json.tool
	@echo "Alert Service:" && curl -s http://localhost:8002/health | python3 -m json.tool

clean: ## コンテナ・ローカルイメージ・ビルド生成物を削除する
	docker compose down -v --rmi local
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	rm -rf alert-service/node_modules alert-service/dist
	rm -f health-checker/health-checker
