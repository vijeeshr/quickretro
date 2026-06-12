FROM node:24.16.0-alpine3.24 AS frontend-builder
WORKDIR /app
# node_modules directory is excluded with .dockerignore
# Copy package files first for efficient caching
COPY src/frontend/package*.json ./
# RUN npm install
RUN npm ci
# Copy source code and run the development build
COPY src/frontend/ .
RUN npm run build-dev

FROM golang:1.26.4-alpine3.24 AS backend-builder
WORKDIR /app
# Copy Go module files and download dependencies
COPY src/go.mod src/go.sum ./
RUN go mod download
# Copy application source code and config
COPY src/config.toml .
COPY src/*.go ./
# Copy compiled frontend assets from the previous stage
COPY --from=frontend-builder /app/dist frontend/dist
# CGO_ENABLED=0 ensures a static binary
# -ldflags "-s -w" removes debugging symbols, reducing binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o retroapp .

FROM alpine:3.24 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch AS final
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
# Copy files and ensure they are owned by the non-root user
COPY --from=backend-builder --chown=10001:10001 /app/retroapp .
COPY --from=backend-builder --chown=10001:10001 /app/config.toml .

# Switch to the non-root user
USER 10001:10001

EXPOSE 8921
CMD ["./retroapp"]