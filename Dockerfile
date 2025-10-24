FROM golang:1.23.5-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o scan-to-nextcloud .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/scan-to-nextcloud .
CMD ["./scan-to-nextcloud"]
