# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Dependencies first, so this layer stays cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# The HTTP API server. Database migrations are applied on startup by main.go.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/smsgateway .

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/smsgateway ./smsgateway

# Migration files, read at startup from this relative path.
COPY --from=builder /src/repository/mysql/migrations ./repository/mysql/migrations

USER app

EXPOSE 8080

ENTRYPOINT ["./smsgateway"]
