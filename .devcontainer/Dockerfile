FROM alpine:3.23.3 AS fetcher

WORKDIR /fetcher

RUN \
  apk add --quiet curl unzip && \
  # Fetch dprint
  curl -fsSL https://dprint.dev/install.sh | sh && \
  mv /root/.dprint/bin/dprint ./dprint && chmod +x ./dprint && \
  # Fetch task
  sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /fetcher v3.48.0 && \
  chmod +x ./task

# Use golang as the base image
FROM golang:1.26-trixie

# Install Node.js
COPY --from=node:24.14.0-trixie /usr/local/ /usr/local/

# Install golangci-lint
COPY --from=golangci/golangci-lint:v2.10.1 /usr/bin/golangci-lint /usr/local/bin/golangci-lint

# Install dprint and task
COPY --from=fetcher /fetcher/dprint /usr/local/bin/dprint
COPY --from=fetcher /fetcher/task /usr/local/bin/task

# Set environment variables
ENV DEBIAN_FRONTEND="noninteractive"

RUN set -e && \
    # Update and install system dependencies
    apt-get update -qq && \
    apt-get install -yqq --no-install-recommends \
        ca-certificates wget curl zip unzip p7zip-full tzdata git tree ripgrep && \
    # Final setup: git config, and cleanup
    git config --global --add safe.directory '*' && \
    cd / && rm -rf /tmp/bin-downloads /var/lib/apt/lists/*

WORKDIR /app

# Add the startup script on every bash session
RUN rm -f /root/.bashrc
COPY .devcontainer/.bashrc /root/.bashrc

# Command just to keep the container running
CMD ["sleep", "infinity"]
