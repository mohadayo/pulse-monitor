.PHONY: up down build test test-python test-go test-ts lint clean

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-python test-go test-ts

test-python:
	cd api-gateway && pip install -r requirements.txt -q && pytest -v

test-go:
	cd health-checker && go test ./... -v

test-ts:
	cd alert-service && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd api-gateway && flake8 app/ tests/ --max-line-length=120

lint-go:
	cd health-checker && go vet ./...

lint-ts:
	cd alert-service && npm run lint

logs:
	docker compose logs -f

health:
	@echo "API Gateway:" && curl -s http://localhost:8000/health | python3 -m json.tool
	@echo "Health Checker:" && curl -s http://localhost:8001/health | python3 -m json.tool
	@echo "Alert Service:" && curl -s http://localhost:8002/health | python3 -m json.tool

clean:
	docker compose down -v --rmi local
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	rm -rf alert-service/node_modules alert-service/dist
	rm -f health-checker/health-checker
