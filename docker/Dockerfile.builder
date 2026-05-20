# =============================================================================
# Dockerfile.builder — Shared Multi-Arch Builder Image
# =============================================================================
#
# This image bundles every build time dependency required by the project and is
# reused by the devcontainer, CI pipeline, and other containers.
#
# Rebuilds only when tool versions or system packages change.
# Published to the github container registry as:
#
#   ghcr.io/varavelio/vdl-builder:sha-<short-commit>
#
# -----------------------------------------------------------------------------
# Usage
# -----------------------------------------------------------------------------
#
# Refer to this image in downstream Dockerfiles:
#
#   FROM ghcr.io/varavelio/vdl-builder:sha-<tag> AS builder
#
# After bumping tool versions in this file, push to main, the pipeline
# publishes the new image automatically. You may then update the tag in
# every downstream Dockerfile manually when ready.
#
# You can also use workflow_dispatch in github actions to trigger a manual
# build and publish of this image, which is useful for testing changes before
# merging to main.
# =============================================================================

# -- fetcher ------------------------------------------------------------------
# Downloads all CLI tool binaries for the target architecture.
# Uses Alpine for a minimal fetch footprint.
# -----------------------------------------------------------------------------

FROM alpine:3.23.3 AS fetcher

WORKDIR /fetcher

RUN \
    apk add --quiet curl unzip && \
    # Fetch task
    sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /fetcher v3.50.0 && \
    chmod +x ./task

# -- final --------------------------------------------------------------------
# Reusable builder image.  Includes Go, Node.js, linting tools, and every
# project CLI tool pre-installed.
# -----------------------------------------------------------------------------

# Use golang as the base image
FROM golang:1.26-trixie

# Install Node.js
COPY --from=node:24.15.0-trixie /usr/local/ /usr/local/

# Install golangci-lint
COPY --from=golangci/golangci-lint:v2.12.2 /usr/bin/golangci-lint /usr/local/bin/golangci-lint

# Install tools from fetcher
COPY --from=fetcher /fetcher/task /usr/local/bin/task

# Set environment variables
ENV DEBIAN_FRONTEND="noninteractive"

RUN set -e && \
    # Update and install system dependencies
    apt-get update -qq && \
    apt-get install -yqq --no-install-recommends \
    ca-certificates wget curl zip unzip p7zip-full tzdata git tree ripgrep \
    python3 python3-pip && \
    # Install zensical
    pip3 install --no-cache-dir --break-system-packages -qq zensical && \
    # Final setup: git config, and cleanup
    git config --global --add safe.directory '*' && \
    cd / && rm -rf /tmp/bin-downloads /var/lib/apt/lists/*

WORKDIR /app

# Command just to keep the container running (useful for dev containers)
CMD ["sleep", "infinity"]
