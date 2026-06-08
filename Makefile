.PHONY: up down dev build logs migrate test seed load-up load-test load-test-100 load-test-1000 load-test-ramp load-test-login ws-test ws-test-100 ws-test-1000 simulate simulate-quick

up:
	docker compose up --build -d

down:
	docker compose down

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

build:
	docker compose build

logs:
	docker compose logs -f

migrate:
	docker compose exec backend /app/server migrate

seed:
	docker compose exec backend /app/server seed

test:
	docker run --rm -v "$(CURDIR)/backend:/app" -w /app golang:1.23-alpine sh -c "go mod download && go test ./..."

COMPOSE_LOAD = docker compose -f docker-compose.yml -f docker-compose.load.yml --profile loadtest

load-up:
	$(COMPOSE_LOAD) up --build -d postgres backend

load-test:
	$(COMPOSE_LOAD) run --rm k6 run /scripts/http-flow.js -e SCENARIO=smoke

load-test-100:
	$(COMPOSE_LOAD) run --rm k6 run /scripts/http-flow.js -e SCENARIO=load_100

load-test-1000:
	$(COMPOSE_LOAD) run --rm k6 run /scripts/http-flow.js -e SCENARIO=load_1000

load-test-ramp:
	$(COMPOSE_LOAD) run --rm k6 run /scripts/http-flow.js -e SCENARIO=ramp_up

load-test-login:
	$(COMPOSE_LOAD) run --rm k6 run /scripts/login.js -e SCENARIO=load_100

ws-test:
	$(COMPOSE_LOAD) run --rm --build wsbots -bots 10 -duration 30s -ramp 5s

ws-test-100:
	$(COMPOSE_LOAD) run --rm --build wsbots -bots 100

ws-test-1000:
	$(COMPOSE_LOAD) run --rm --build wsbots -bots 1000 -ramp 2m -duration 5m

simulate:
	$(COMPOSE_LOAD) run --rm --build simulator

simulate-quick:
	$(COMPOSE_LOAD) run --rm --build simulator -users 20 -conversations 5 -messages-per-second 10 -duration 30s -ramp 5s
