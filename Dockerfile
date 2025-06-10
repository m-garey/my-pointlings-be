FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache make git

# Copy go.mod, go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /pointlings-backend ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /pointlings-backend /pointlings-backend
COPY .env.example /.env

EXPOSE 8080

ENTRYPOINT ["/pointlings-backend"]
