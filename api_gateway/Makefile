APP_NAME=api-gateway
MAIN_PATH=cmd/app/main.go

.PHONY: run gen-auth swagger-gen

run:
	go run $(MAIN_PATH)

gen-auth:
	protoc -I api/proto api/proto/auth.proto \
	--go_out=pkg/api/auth/ --go_opt=paths=source_relative \
	--go-grpc_out=pkg/api/auth/ --go-grpc_opt=paths=source_relative

swagger-gen:
	swag init -g cmd/app/main.go
