FROM golang:1.24-alpine3.23

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o worker-service cmd/main.go

EXPOSE 8081
CMD ["./worker-service"]