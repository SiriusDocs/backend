# Sirius docs backend project

# Развёртывание проекта (локально)
(пока не упаковывал в докер, извините)

1. Убедиться, что в корне директории каждого микросервиса существует файл конфига по пути `{microservice}/configs/server/local.yaml`

Для `api_gateway`:

`local.yaml`
```yaml
env: "local"

http_server:
  host: "localhost"
  port: 8080
  timeout: 4s
  idle_timeout: 60s
  allowed_origins:
  - http://localhost:5173
  - http://localhost:5174
  - http://localhost:3000


clients:
  auth_service:
    address: "localhost:44044"
    timeout: 5s
    retries_count: 3

  temp_service:
    address: "localhost:44045"
    timeout: 5s
    retries_count: 3
```

для `auth_service`:

`local.yaml`
```yaml
auth_service:
  port: 44044
  timeout: 5s
  retries_count: 3
```

для `template_service`:

`local.yaml`
```yaml
temp_service:
  port: 44045
  timeout: 5s
  retries_count: 3
```

2. Убедиться, что в корне директории каждого микросервиса существует файл `.env`

Для `api_gateway`:
```
CONFIG_PATH=./configs/server/local.yaml
SIGNING_KEY={secret_key_in_auth_service}
```

для `auth_service`:
```
CONFIG_PATH=./configs/server/local.yaml
MIGRATIONS_PATH=./migrations
DB_HOST=localhost
DB_PASSWORD=goida228
DB_USERNAME=wolkodaf
DB_PORT=5432
DB_NAME=postgres
DB_SSL=disable

SIGNING_KEY={secret_key_in_auth_service}
SALT={random_symbols}
```

**!!!Важно!!! SIGNING_KEY из auth_service и из api_gateway должен совпадать**

для `template_service`:
```
CONFIG_PATH=./configs/server/local.yaml
MIGRATIONS_PATH=./migrations
DB_HOST=localhost
DB_PASSWORD=goida228
DB_USERNAME=wolkodaf
DB_PORT=5432
DB_NAME=postgres
DB_SSL=disable
```

3. Запущен docker-контейнер с базой данных PostgreSQL

Команда для запуска:
```bash
docker run -d -p 5432:5432 --name postgres -e POSTGRES_USER=wolkodaf -e POSTGRES_PASSWORD=goida228 postgres:17.6
```

4. Запуск приложения:

    1.  В директории `template_service`:
    ```bash
    go mod download
    go run cmd/migrator/main.go -db-url="postgres://wolkodaf:goida228@localhost:5432/postgres?sslmode=disable" -migrations-path=./migrations -migrations-table=temp --level=up
    go run cmd/temp/main.go 
    ```

    2. В директории `auth_service`:
    
    Запуск мигратора ИНДИВИДУАЛЕН для каждого микросервиса
    ```bash
    go mod download
    go run cmd/migrator/main.go -db-url="postgres://wolkodaf:goida228@localhost:5432/postgres?sslmode=disable" -migrations-path=./migrations -migrations-table=auth --level=up
    go run cmd/auth_user/main.go
    ```

    3. В директории `api_gateway`:

    ```bash
    go mod download
    go run cmd/app/main.go
    ```

**Создание дефолтного админа**

Также для управления системой нужно создать админа такой командой (из auth_service)
```bash
go run ./cmd/create_admin/main.go -u superadmin -e myadmin@docs.ru -p goida1234
```
`-env` - расположение .env файла

`-u` - username админа

`-e` - email админа

`-p` - password админа
