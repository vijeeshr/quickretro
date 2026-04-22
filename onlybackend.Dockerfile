# docker build -f onlybackend.Dockerfile -t quickretro-app .

FROM golang:1.26.2-alpine3.23 AS backend-builder
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

FROM alpine:3.23 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch AS final
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
# Copy files and ensure they are owned by the non-root user
COPY --from=backend-builder --chown=10001:10001 /app/retroapp .
COPY --from=backend-builder --chown=10001:10001 /app/config.toml .

# Switch to the non-root user
USER 10001:10001

EXPOSE 8080
CMD ["./retroapp"]