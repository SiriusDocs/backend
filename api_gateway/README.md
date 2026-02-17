# Api-шлюз для маршрутизации запросов

Пример config-файла (local):
`./configs/server/local.yaml`
```yaml
env: "local"

http_server:
  host: "localhost"
  port: 8080
  timeout: 4s
  idle_timeout: 60s

clients:
  auth_service:
    address: "localhost:44044"
    timeout: 5s
    retries_count: 3

 ...
```

Запуск микросервиса (без docker):
```bash
go mod download
go run cmd/app/main.go
```

Swagger документация доступна по адресу:
`http://localhost:8080/swagger/index.html`


