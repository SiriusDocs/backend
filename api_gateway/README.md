# Api-шлюз для маршрутизации запросов

**Пример config-файла (local):**
`./configs/server/local.yaml`
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

 ...
```
**Переменные окружения:**
В корне проекта создать .env:
```
CONFIG_PATH=./configs/server/local.yaml 
```

**Запуск микросервиса (без docker):**
```bash
go mod download
go run cmd/app/main.go
```

Swagger документация доступна по адресу:
`http://localhost:8080/swagger/index.html`


