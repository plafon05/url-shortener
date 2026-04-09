# URL Shortener (Go + PostgreSQL)

Сервис сокращения ссылок на Go с хранением в PostgreSQL.

## Возможности

- Создание короткого alias для URL (с ручным alias или автогенерацией).
- Редирект по alias на исходный URL.
- Получение всех alias по исходному URL.
- Удаление записи по alias.
- Логирование через `slog`.
- Запуск локально и в Docker.

## Стек

- Go `1.24`
- PostgreSQL `17`
- Роутер: `chi`
- Тесты: `go test`, `httpexpect`, `testify`

## Структура проекта

```text
cmd/url-shortener/                # entrypoint приложения
internal/config/                  # загрузка YAML-конфига
internal/http-server/handlers/    # HTTP-обработчики
internal/storage/postgres/        # реализация хранилища на PostgreSQL
config/                           # окружения (local/prod)
tests/                            # интеграционные тесты
```

## Конфигурация

Приложение читает путь к конфигу из переменной окружения:

```bash
CONFIG_PATH=./config/local.yaml
```

Поддерживаемые поля в YAML:

```yaml
env: "local" # local | dev | prod

postgres:
  dsn: "host=localhost port=5433 user=postgres password=1230 dbname=url_shortener sslmode=disable"

http_server:
  address: "localhost:8082"
  timeout: 4s
  idle_timeout: 60s
  user: "myuser"
  password: "mypass"
```

## Быстрый старт (локально)

1. Поднять PostgreSQL:

```bash
docker compose -f docker-compose-local.yml up -d
```

2. Запустить сервис:

```bash
export CONFIG_PATH=./config/local.yaml
go run ./cmd/url-shortener
```

Сервис будет доступен на `http://localhost:8082`.

## Запуск в Docker Compose

Запуск:

```bash
docker compose up -d --build
```

Приложение поднимется на `http://localhost:8082`.

### Production-конфиг (`config/prod.yaml`)

В `prod.yaml` секреты не хранятся в явном виде, а подставляются из переменных окружения:

```yaml
env: "prod"

postgres:
  dsn: "host=postgres port=5432 user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=url_shortener sslmode=disable"

http_server:
  address: "0.0.0.0:8082"
  timeout: 4s
  idle_timeout: 60s
  user: "${HTTP_USER}"
  password: "${HTTP_PASSWORD}"
```

Для деплоя через GitHub Actions значения `POSTGRES_USER`, `POSTGRES_PASSWORD`, `HTTP_USER`, `HTTP_PASSWORD` должны приходить из GitHub Secrets.

## API

Все методы под `/url` требуют Basic Auth (локально можно использовать `config/local.yaml`, в production значения берутся из env).

### 1) Создать короткую ссылку

`POST /url/`

Пример:

```bash
curl -X POST http://localhost:8082/url/ \
  -u myuser:mypass \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/some/path","alias":"my-alias"}'
```

Без `alias` сервис сгенерирует его автоматически:

```bash
curl -X POST http://localhost:8082/url/ \
  -u myuser:mypass \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/some/path"}'
```

Успешный ответ:

```json
{
  "status": "OK",
  "alias": "my-alias"
}
```

### 2) Редирект по alias

`GET /{alias}`

Пример:

```bash
curl -i http://localhost:8082/my-alias
```

Ответ: `302 Found` с заголовком `Location`.

### 3) Получить alias по URL

`GET /url/aliases?url=<encoded_url>`

Пример:

```bash
curl "http://localhost:8082/url/aliases?url=https%3A%2F%2Fexample.com%2Fsome%2Fpath" \
  -u myuser:mypass
```

Успешный ответ:

```json
{
  "status": "OK",
  "alias": ["my-alias", "another-alias"]
}
```

### 4) Удалить alias

`DELETE /url/{alias}`

Пример:

```bash
curl -X DELETE http://localhost:8082/url/my-alias \
  -u myuser:mypass
```

Успешный ответ:

```json
{
  "status": "OK"
}
```

## Makefile команды

```bash
make up         # docker compose up -d
make down       # docker compose down
make restart    # перезапуск контейнеров
make logs       # логи postgres
make psql       # psql в контейнере my_postgres (для локального compose)
make run        # go run ./cmd/url-shortener
make build      # сборка бинарника в bin/url-shortener
make test       # go test ./...
make fmt        # go fmt ./...
make lint       # golangci-lint run
make clean-db   # удалить docker volume с данными postgres
```

## Тесты

```bash
go test ./...
```

Интеграционные тесты находятся в `tests/url_shortener_test.go` и ожидают, что сервис доступен на `localhost:8082`.

## Примечания

- Таблица `url` и индекс по alias создаются автоматически при старте приложения.
- Для production секреты не должны храниться в репозитории: используйте только переменные окружения/GitHub Secrets.
