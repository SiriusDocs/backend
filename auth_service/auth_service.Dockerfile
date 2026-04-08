FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build cmd/auth_user/main.go -o ./auth_service.go

CMD ["/app/auth_service.go"]