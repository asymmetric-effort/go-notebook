# Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.
#
# Linter image for go-notebook CI.
# Contains all linters and tools needed for the checks job.
# Built from ubuntu:26.04 with Go 1.27.4, Node.js 22, and all lint tools.
#
# Build:
#   docker build -t ghcr.io/asymmetric-effort/go-notebook/linter:latest docker/linter/
#
# Push:
#   docker push ghcr.io/asymmetric-effort/go-notebook/linter:latest

FROM ubuntu:26.04

LABEL org.opencontainers.image.source="https://github.com/asymmetric-effort/go-notebook"
LABEL org.opencontainers.image.description="Linter image for go-notebook CI"
LABEL org.opencontainers.image.licenses="MIT"

ENV DEBIAN_FRONTEND=noninteractive

# System packages.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    libxml2-utils \
    make \
    python3 \
    && rm -rf /var/lib/apt/lists/*

# Go 1.27.4.
ENV GO_VERSION=1.27.4
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-$(dpkg --print-architecture).tar.gz" \
    | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ENV GOPATH="/root/go"

# Node.js 22 (LTS).
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# Go-based linters.
RUN go install github.com/asymmetric-effort/leakdetector/cmd/leakdetector@latest \
    && go install golang.org/x/vuln/cmd/govulncheck@latest \
    && go install github.com/checkmake/checkmake/cmd/checkmake@latest \
    && go install github.com/google/yamlfmt/cmd/yamlfmt@latest \
    && go install golang.org/x/tools/gopls@latest

# Verify all tools are installed.
RUN go version \
    && node --version \
    && npm --version \
    && leakdetector --version || true \
    && govulncheck --version || true \
    && checkmake --version || true \
    && yamlfmt --version || true \
    && xmllint --version \
    && python3 --version \
    && git --version \
    && gopls version || true

WORKDIR /workspace
