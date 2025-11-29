# ---------------------------
# Variables
# ---------------------------
APP_NAME=url-shortener
DOCKER_COMPOSE=docker compose
GO=go

# ---------------------------
# Docker commands
# ---------------------------

# Запустить PostgreSQL (docker-compose)
up:
	$(DOCKER_COMPOSE) up -d

# Остановить контейнеры
down:
	$(DOCKER_COMPOSE) down

# Перезапустить контейнеры
restart:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d

# Посмотреть логи PostgreSQL
logs:
	$(DOCKER_COMPOSE) logs -f postgres

# Зайти в psql (внутрь контейнера)
psql:
	docker exec -it my_postgres psql -U postgres -d url-shortener

# ---------------------------
# Go commands
# ---------------------------

# Запустить приложение
run:
	$(GO) run ./cmd/$(APP_NAME)

# Собрать бинарник
build:
	$(GO) build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

# Запустить тесты
test:
	$(GO) test ./...

# Форматировать код
fmt:
	$(GO) fmt ./...

# Проверить линтером
lint:
	golangci-lint run

# ---------------------------
# Полезные команды
# ---------------------------

# Удалить volume PostgreSQL
clean-db:
	docker volume rm httpserver_project_pgdata || true
