FROM golang:1.24-alpine3.23

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o agent-service ./cmd/main.go

CMD ["./agent-service"]