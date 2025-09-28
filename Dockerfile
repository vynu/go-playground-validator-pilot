# Optimized Dockerfile for Go 1.25.1 Playground Validator
# Follows 2025 best practices for static Go binaries with distroless runtime
#
# Build Arguments (configurable):
#   GOMEMLIMIT_ARG: Go memory limit (default: 512MiB)
#   GOGC_ARG: Garbage collection target percentage (default: 100)
#   GOMAXPROCS_ARG: Maximum number of OS threads (default: 0 = auto)
#
# Usage Examples:
#   # Default build
#   docker build -t validator .
#
#   # Custom memory limit
#   docker build --build-arg GOMEMLIMIT_ARG=1GiB -t validator .
#
#   # Custom GC settings
#   docker build --build-arg GOGC_ARG=50 --build-arg GOMEMLIMIT_ARG=2GiB -t validator .
#
#   # Target specific stage
#   docker build --target distroless --build-arg GOMEMLIMIT_ARG=1GiB -t validator:distroless .
#   docker build --target alpine --build-arg GOMEMLIMIT_ARG=1GiB -t validator:alpine .

# ===========================
# Stage 1: Build Environment
# ===========================
FROM golang:1.25-alpine AS builder

# Build metadata
LABEL stage="builder"
LABEL description="Optimized build stage for Go 1.25.1 Playground Validator"

# Essential build dependencies for static compilation
RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    file \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Critical environment variables for optimal static binary compilation
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    GOSUMDB=sum.golang.org

# Copy dependency files first for optimal Docker layer caching
COPY src/go.mod src/go.sum ./

# Download and verify dependencies (cached layer)
RUN go mod download && go mod verify

# Copy source code
COPY src/ .

# Build optimized static binary using 2025 best practices
# -trimpath: removes file system paths for reproducible builds
# -ldflags "-s -w": strips symbol table and debug info
# -a: force rebuild of packages
# -installsuffix cgo: separate package cache for CGO_ENABLED=0
RUN go build \
    -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -a \
    -installsuffix cgo \
    -o validator \
    main.go

# Verify the binary is properly compiled and static
RUN file validator && \
    (ldd validator 2>&1 | grep -q "not a dynamic executable" || file validator | grep -q "statically linked") && \
    chmod +x validator && \
    ls -lah validator

# ===========================
# Stage 2: Distroless Runtime
# ===========================
FROM gcr.io/distroless/static-debian12:nonroot AS distroless

# Build arguments for runtime configuration
ARG GOMEMLIMIT_ARG=512MiB
ARG GOGC_ARG=100
ARG GOMAXPROCS_ARG=0

# Runtime metadata optimized for 2025
LABEL maintainer="Go Playground Validator Team" \
      description="Ultra-minimal Go 1.25.1 validation server" \
      version="2.0.0-go1.25.1" \
      org.opencontainers.image.title="Go Playground Validator" \
      org.opencontainers.image.description="Modular validation server with automatic model discovery" \
      org.opencontainers.image.version="2.0.0" \
      org.opencontainers.image.vendor="Go Playground Validator" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.base.name="gcr.io/distroless/static-debian12:nonroot" \
      org.opencontainers.image.source="https://github.com/your-org/go-playground-validator"

# Copy essential system files from builder for Go runtime
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the optimized static binary
COPY --from=builder /app/validator /validator

# Copy source directories needed for dynamic model discovery
COPY --from=builder /app/models /src/models
COPY --from=builder /app/validations /src/validations

# Optimized runtime environment variables for Go 1.25.1
# GOMEMLIMIT: Controls Go memory usage (e.g., 512MiB, 1GiB, 2GiB)
# GOGC: Garbage collection target percentage (default: 100)
# GOMAXPROCS: Maximum number of OS threads (0 = use all available CPUs)
ENV PORT=8080 \
    TZ=UTC \
    GOGC=${GOGC_ARG} \
    GOMEMLIMIT=${GOMEMLIMIT_ARG} \
    GOTRACEBACK=none \
    GOMAXPROCS=${GOMAXPROCS_ARG}

# Expose application port
EXPOSE 8080

# Use non-root user (uid:gid 65532:65532) for security
USER nonroot:nonroot

# Set working directory to root so relative paths work
WORKDIR /

# Set entrypoint in vector form (required for distroless)
ENTRYPOINT ["/validator"]

# Optional command arguments
CMD []

# ===========================
# Stage 3: Alpine Runtime (Alternative)
# ===========================
FROM alpine:3.19 AS alpine

# Build arguments for runtime configuration (reuse from distroless stage)
ARG GOMEMLIMIT_ARG=512MiB
ARG GOGC_ARG=100
ARG GOMAXPROCS_ARG=0

# Alpine metadata
LABEL description="Alpine-based Go 1.25.1 validation server with debugging tools" \
      version="2.0.0-alpine" \
      org.opencontainers.image.base.name="alpine:3.19"

# Install minimal runtime dependencies and debugging tools
RUN apk update && apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    wget \
    && addgroup -g 1001 -S nonroot \
    && adduser -u 1001 -S nonroot -G nonroot \
    && rm -rf /var/cache/apk/*

# Copy the optimized static binary from builder
COPY --from=builder /app/validator /validator

# Copy source directories needed for dynamic model discovery
COPY --from=builder /app/models /src/models
COPY --from=builder /app/validations /src/validations

# Set proper ownership
RUN chown nonroot:nonroot /validator && chmod +x /validator

# Runtime environment variables (configurable via build args)
# GOMEMLIMIT: Controls Go memory usage (e.g., 512MiB, 1GiB, 2GiB)
# GOGC: Garbage collection target percentage (default: 100)
# GOMAXPROCS: Maximum number of OS threads (0 = use all available CPUs)
ENV PORT=8080 \
    TZ=UTC \
    GOGC=${GOGC_ARG} \
    GOMEMLIMIT=${GOMEMLIMIT_ARG} \
    GOMAXPROCS=${GOMAXPROCS_ARG}

# Switch to non-root user
USER nonroot:nonroot

# Set working directory to root so relative paths work
WORKDIR /

# Expose application port
EXPOSE 8080

# Health check with wget (available in Alpine)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["/validator"]
CMD []