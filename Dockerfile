# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o goerp cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/goerp .
COPY --from=builder /app/config ./config
COPY --from=builder /app/apps ./apps
EXPOSE 8080
CMD ["./goerp"]
