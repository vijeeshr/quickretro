# docker build -f onlybackend.Dockerfile -t quickretro-app .

FROM golang:1.22.0-alpine AS builder
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
# COPY src/ .
COPY src/*.go .
COPY src/frontend/dist frontend/dist
COPY src/config.toml .
RUN CGO_ENABLED=0 GOOS=linux go build -o retroapp .

FROM alpine:latest AS certificates
RUN apk --no-cache add ca-certificates

FROM scratch
ENV REDIS_HOST redis:6379
WORKDIR /app
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/retroapp .
COPY --from=builder /app/config.toml .
EXPOSE 8080
CMD ["./retroapp"]