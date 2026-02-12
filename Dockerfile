FROM node:24.11.1-alpine AS frontend-builder
WORKDIR /app
# node_modules directory is excluded with .dockerignore
# Copy package files first for efficient caching
COPY src/frontend/package*.json ./
RUN npm install
# Copy source code and run the development build
COPY src/frontend/ .
RUN npm run build-dev

FROM golang:1.26.0-alpine AS backend-builder
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

FROM scratch AS final
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY --from=backend-builder /app/retroapp .
COPY --from=backend-builder /app/config.toml .

EXPOSE 8080
CMD ["./retroapp"]