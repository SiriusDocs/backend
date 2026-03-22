# Сервис авторизации
Готова логика регистрации, аутентификации (с refresh токеном)

**Пример config-файла (local):**
`./configs/server/server.yaml`
```yaml
auth_service:
  port: 44044
  timeout: 5s
  retries_count: 3
```

**Переменные окружения:**
В корне проекта создать .env:
```
CONFIG_PATH=./configs/server/server.yaml
MIGRATIONS_PATH=./migrations

DB_HOST=localhost
DB_PASSWORD=goida228
DB_USERNAME=wolkodaf
DB_PORT=5432
DB_NAME=postgres
DB_SSL=disable

SIGNING_KEY=<random_symbols>
SALT=<random_symbols>
```
Пример:
```bash
docker run -d -p 5432:5432 --name postgres -e POSTGRES_USER=wolkodaf -e POSTGRES_PASSWORD=goida228 postgres:17.6
```

**Запуск микросервиса (без docker):**
```bash
go mod download
go run cmd/migrator/main.go -db-url="postgres://wolkodaf:goida228@localhost:5432/postgres?sslmode=disable" -migrations-path=./migrations --level=up
go run cmd/auth_user/main.go
```
