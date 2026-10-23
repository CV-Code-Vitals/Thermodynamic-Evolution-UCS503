# Pinned, hardened sandbox container for untrusted candidate code evaluation
# UCS503 Thermodynamic Architecture Evolution — Milestone 0
FROM golang:1.25.0-alpine3.21

# Install minimal tools required for go test
RUN apk --no-cache add git

# Create dedicated non-root runner user with explicit UID/GID
RUN addgroup -g 10001 runner && \
    adduser -u 10001 -G runner -s /bin/sh -D runner

# Setup temporary and workspace directories
RUN mkdir -p /workspace /tmp/gocache /tmp/go && \
    chown -R runner:runner /workspace /tmp/gocache /tmp/go

# Hardened environment defaults:
# - CGO disabled to prevent native execution vulnerabilities
# - Dedicated temporary caches in tmpfs
ENV CGO_ENABLED=0 \
    GOCACHE=/tmp/gocache \
    GOPATH=/tmp/go

WORKDIR /workspace
USER 10001:10001

CMD ["go", "test", "-v", "./..."]
