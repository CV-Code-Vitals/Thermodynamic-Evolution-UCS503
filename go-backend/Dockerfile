# -----------------------------------------------------------------------------
# Stage 1: Build Rust AST Engine
# -----------------------------------------------------------------------------
FROM rust:alpine AS rust-builder
WORKDIR /rust-app
RUN apk add --no-cache musl-dev
COPY thermodynamic-ast-engine/ ./
RUN cargo build --release --locked

# -----------------------------------------------------------------------------
# Stage 2: Build Go Backend
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS go-builder
WORKDIR /go-app
COPY go-backend/go.mod go-backend/go.sum ./
RUN go mod download
COPY go-backend/ ./
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -o backend_server .

# -----------------------------------------------------------------------------
# Stage 3: Minimal Production Runtime
# -----------------------------------------------------------------------------
FROM alpine:latest
RUN apk --no-cache add ca-certificates git docker-cli wget
WORKDIR /app

# Create unprivileged application user and group
RUN addgroup -g 10002 appgroup && \
    adduser -u 10002 -G appgroup -s /bin/sh -D appuser

# Copy compiled artifacts
COPY --from=go-builder /go-app/backend_server .
COPY --from=rust-builder /rust-app/target/release/thermodynamic-ast-engine .
RUN chmod +x thermodynamic-ast-engine backend_server

# Create storage directories for uploads, deliverables, and sandboxes
RUN mkdir -p /app/uploads /app/sandboxes && \
    chown -R appuser:appgroup /app/uploads /app/sandboxes && \
    chmod 750 /app/uploads /app/sandboxes

EXPOSE 8080 8082
ENV ENGINE_PATH=/app/thermodynamic-ast-engine
USER appuser:appgroup
CMD ["./backend_server"]
