FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /holodeck-art-api

FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /holodeck-art-api .

EXPOSE 8080

CMD ["./holodeck-art-api"]

