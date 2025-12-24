FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go mod init log-generator && \
    go get github.com/google/uuid && \
    CGO_ENABLED=0 go build -o log-generator .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/log-generator .

ENV LOG_DIR=/var/log/app
ENV LOG_INTERVAL=2s

CMD ["./log-generator"]
