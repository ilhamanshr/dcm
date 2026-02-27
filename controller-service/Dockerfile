FROM golang:1.24-alpine3.23

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o controller-service ./cmd/main.go

EXPOSE 8080
CMD ["./controller-service"]