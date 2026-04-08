FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build cmd/app/main.go -o ./api_gateway

CMD ["/app/api_gateway"]