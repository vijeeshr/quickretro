FROM --platform=$BUILDPLATFORM node:24.15.0-alpine3.23 AS frontend-builder
WORKDIR /app
# node_modules directory is excluded with .dockerignore
# Copy package files first for efficient caching
COPY src/frontend/package*.json ./
# RUN npm install
RUN npm ci
# Copy source code and run the build
COPY src/frontend/ .
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26.3-alpine3.23 AS backend-builder
# TARGETOS and TARGETARCH are automatically set by Docker Buildx
# Using --platform=$BUILDPLATFORM runs the Go compiler natively (fast),
# then cross-compiles to the target via GOOS/GOARCH (instead of running
# the entire Go toolchain under slow QEMU emulation).
ARG TARGETOS TARGETARCH
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
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags netgo,osusergo -ldflags "-s -w" -o retroapp .

FROM alpine:3.23 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch AS final
LABEL org.opencontainers.image.source="https://github.com/vijeeshr/quickretro"
LABEL org.opencontainers.image.description="QuickRetro is a free and open-source sprint retrospective app for agile teams. Self-hosted, real-time, mobile-friendly, and privacy-first."
LABEL org.opencontainers.image.licenses="AGPL-3.0"
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
# Copy files and ensure they are owned by the non-root user
COPY --from=backend-builder --chown=10001:10001 /app/retroapp .
COPY --from=backend-builder --chown=10001:10001 /app/config.toml .

# Switch to the non-root user
USER 10001:10001

EXPOSE 8921
CMD ["./retroapp"]