FROM node:22.14.0-alpine AS frontend-builder
WORKDIR /app
# node_modules directory is excluded with .dockerignore
COPY src/frontend/ .
RUN npm install
RUN npm run build-dev

FROM golang:1.24.2-alpine AS backend-builder
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
# COPY src/ .
COPY src/*.go ./
# COPY src/frontend/dist frontend/dist
COPY src/config.toml .
COPY --from=frontend-builder /app/dist frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o retroapp .

FROM alpine:latest AS certificates
RUN apk --no-cache add ca-certificates

FROM scratch
WORKDIR /app
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend-builder /app/retroapp .
COPY --from=backend-builder /app/config.toml .
# COPY src/public ./public
EXPOSE 8080
CMD ["./retroapp"]