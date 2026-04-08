FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build cmd/migrator/main.go -o ./migrator.go

CMD ["/app/migrator.go"]