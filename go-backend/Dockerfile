# -----------------------------------------------------------------------------
# Stage 1: Build Rust AST Engine
# -----------------------------------------------------------------------------
FROM rust:alpine AS rust-builder
WORKDIR /rust-app
RUN apk add --no-cache musl-dev
COPY thermodynamic-ast-engine/ ./
RUN cargo build --release

# -----------------------------------------------------------------------------
# Stage 2: Build Go Backend
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS go-builder
WORKDIR /go-app
COPY go-backend/go.mod go-backend/go.sum ./
RUN go mod download
COPY go-backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend_server .

# -----------------------------------------------------------------------------
# Stage 3: Minimal Production Runtime
# -----------------------------------------------------------------------------
FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /app

# Copy compiled artifacts
COPY --from=go-builder /go-app/backend_server .
COPY --from=rust-builder /rust-app/target/release/thermodynamic-ast-engine .
RUN chmod +x thermodynamic-ast-engine backend_server

# Create storage directory for uploads and deliverables
RUN mkdir -p /app/uploads && chmod 750 /app/uploads

EXPOSE 8080
ENV ENGINE_PATH=/app/thermodynamic-ast-engine
CMD ["./backend_server"]
