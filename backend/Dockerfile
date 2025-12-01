# Build Stage - NOTE THE 1.24 VERSION TAG
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Only copy go.mod for now (we have no external dependencies yet)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main ./cmd/server/main.go

# Run Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]