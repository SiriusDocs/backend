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

file_service:
  address: "localhost:50054"
  timeout: 5s
```

для `template_service`:

`local.yaml`
```yaml
temp_service:
  port: 44045
  timeout: 5s
  retries_count: 3
```

для `file_service`:

`local.yaml`
```yaml
env: "local"

grpc:
  port: 50054
  timeout: "5s"

s3:
  endpoint: "https://s3.twcstorage.ru"
  region: "ru-1"
  bucket: "..."
  max_upload_size: 524288000     # 500 MiB — потолок размера одного файла (env: S3_MAX_UPLOAD_SIZE)
  upload_url_ttl: 15m            # сколько живёт presigned-ссылка на загрузку
  download_url_ttl: 15m          # сколько живёт presigned-ссылка на скачивание
  allowed_content_types:         # белый список MIME (env: S3_ALLOWED_CONTENT_TYPES, через запятую)
    - image/jpeg
    - image/png
    - image/webp
    - image/heic
    - video/mp4
    - video/quicktime
    - application/pdf

```

2. Убедиться, что в корне директории каждого микросервиса существует файл `.env`

Для `api_gateway`:
```
CONFIG_PATH=./configs/server/local.yaml
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

SIGNING_KEY={random_symbols}
SALT={random_symbols}
```

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

для `file_service`:
```
CONFIG_PATH=./configs/server/local.yaml

DB_PASSWORD=
DB_HOST=localhost

SALT="..."

SIGNING_KEY="..."

SWAGGER_HOST="localhost:8080"

S3_ACCESS_KEY="..."
S3_SECRET_KEY="..."
# Необязательные переопределения yaml (можно оставить пустыми — возьмётся local.yaml):
S3_BUCKET="..."
S3_ENDPOINT="https://s3.twcstorage.ru"
S3_REGION="ru-1"
S3_MAX_UPLOAD_SIZE=524288000
S3_ALLOWED_CONTENT_TYPES="image/jpeg,image/png,video/mp4,application/pdf"
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

    3. В директории `file_service`:

    ```bash
    go mod download
    go run cmd/file/main.go
    ```

    4. В директории `api_gateway`:

    ```bash
    go mod download
    go run cmd/app/main.go
    ```


