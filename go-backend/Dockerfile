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
RUN --mount=type=cache,target=/root/.cache/go-build GOMEMLIMIT=1GiB CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -p 1 -o backend_server .

# Cache dependencies for the bundled repair fixtures at build time. Runtime
# verification is offline and must never fetch modules while processing a demo.
FROM golang:1.25-alpine AS repair-deps
COPY benchmarks/external_v2/fatih_color/go.mod benchmarks/external_v2/fatih_color/go.sum /fixtures/fatih_color/
COPY benchmarks/external_v2/stretchr_objx/go.mod benchmarks/external_v2/stretchr_objx/go.sum /fixtures/stretchr_objx/
COPY benchmarks/external_v2/schollz_progressbar/go.mod benchmarks/external_v2/schollz_progressbar/go.sum /fixtures/schollz_progressbar/
COPY benchmarks/external_v2/gorilla_websocket/go.mod benchmarks/external_v2/gorilla_websocket/go.sum /fixtures/gorilla_websocket/
RUN set -e; for dir in /fixtures/*; do (cd "$dir" && GOMODCACHE=/repair-modcache go mod download); done

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
COPY --from=go-builder /usr/local/go /usr/local/go
COPY --from=repair-deps /repair-modcache /repair-modcache
COPY benchmarks/external_v2/fatih_color /app/benchmarks/external_v2/fatih_color
COPY benchmarks/external_v2/stretchr_objx /app/benchmarks/external_v2/stretchr_objx
COPY benchmarks/external_v2/schollz_progressbar /app/benchmarks/external_v2/schollz_progressbar
COPY benchmarks/external_v2/gorilla_websocket /app/benchmarks/external_v2/gorilla_websocket
COPY benchmarks/heldout_v1/microservice_api /app/benchmarks/heldout_v1/microservice_api
RUN chmod +x thermodynamic-ast-engine backend_server

# Create storage directories for uploads, deliverables, and sandboxes
RUN mkdir -p /app/uploads /app/sandboxes /app/.tmp && \
    chown -R appuser:appgroup /app/uploads /app/sandboxes /app/.tmp && \
    chmod 750 /app/uploads /app/sandboxes /app/.tmp

EXPOSE 8080 8082
ENV ENGINE_PATH=/app/thermodynamic-ast-engine
ENV WORKSPACE_ROOT=/app PATH=/usr/local/go/bin:$PATH GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOCACHE=/tmp/go-build GOMODCACHE=/repair-modcache
USER appuser:appgroup
CMD ["./backend_server"]
