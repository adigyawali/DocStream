# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (cached layers)
COPY go.mod ./
# (We don't have go.sum yet because we haven't added dependencies, so we skip copying it for now)

COPY . .

# Build the Go app
# We point to main.go inside cmd/server
RUN go build -o main ./cmd/server/main.go

# Run Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]