FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build cmd/temp/main.go -o ./template_service.go

CMD ["/app/template_service.go"]