FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o ./auth_service cmd/auth_user/main.go

CMD ["/app/auth_service.go"]