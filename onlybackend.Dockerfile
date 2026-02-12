# docker build -f onlybackend.Dockerfile -t quickretro-app .

FROM golang:1.26.0-alpine AS backend-builder
WORKDIR /app
# Copy Go module files and download dependencies
COPY src/go.mod src/go.sum ./
RUN go mod download
# Copy application source code and config
COPY src/config.toml .
COPY src/*.go ./
# Frontend must be already built by doing "npm run build" or "npm run build-dev" before this step
COPY src/frontend/dist frontend/dist
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