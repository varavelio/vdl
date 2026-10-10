#########
# TOOLS #
#########

# Use golang (debian) as the base image
FROM golang:1.27.2-trixie as tools

# Install node.js
COPY --from=node:26.11.1-trixie-slim /usr/local /usr/local

# Install golangci-lint
COPY --from=golangci/golangci-lint:v2.14.0 /usr/bin/golangci-lint /usr/local/bin/golangci-lint

# Install veta
COPY --from=varavel/veta:0.1.1 /usr/local/bin/veta /usr/local/bin/veta

# Install task
COPY --from=ghcr.io/varavelio/container-tools/task:3.54.0 /usr/local/bin/task /usr/local/bin/task

# Install dprint
COPY --from=ghcr.io/varavelio/container-tools/dprint:0.61.1 /usr/local/bin/dprint /usr/local/bin/dprint

# Environment variables
ENV \
  PIP_BREAK_SYSTEM_PACKAGES=1 \
  CGO_ENABLED=0

# Build time variables
ARG DEBIAN_FRONTEND=noninteractive

RUN set -e && \
  # Install system dependencies
  apt-get update -qq && \
  apt-get install -yqq --no-install-recommends \
  ca-certificates wget curl zip unzip p7zip-full tzdata git tree ripgrep \
  python3 python3-pip && \
  rm -rf /var/lib/apt/lists/* && \
  # Git config
  git config --system --add safe.directory '*'

WORKDIR /workspaces/vdl

################
# DEVCONTAINER #
################

FROM tools AS devcontainer

CMD ["sleep", "infinity"]

###########
# BUILDER #
###########

FROM tools AS builder

# Warm the Go module cache before copying the sources so dependency
# downloads stay cached across source-only changes.
COPY toolchain/go.mod toolchain/go.sum ./toolchain/
RUN cd ./toolchain && go mod download

COPY . .

# Build time variables (here to prevent cache invalidation when this change)
ARG \
  VDL_VERSION \
  VDL_COMMIT \
  VDL_DATE

RUN task build:prod

##############
# PRODUCTION #
##############

FROM debian:trixie-slim AS production

# Build time variables
ARG \
  DEBIAN_FRONTEND=noninteractive \
  VDL_VERSION

LABEL org.opencontainers.image.title="VDL" \
      org.opencontainers.image.authors="Varavel" \
      org.opencontainers.image.description="VDL is an open-source, type-safe, multi-language, and easily extensible schema definition language and code generation toolchain." \
      org.opencontainers.image.url="https://github.com/varavelio/vdl" \
      org.opencontainers.image.documentation="https://github.com/varavelio/vdl" \
      org.opencontainers.image.source="https://github.com/varavelio/vdl" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Varavel" \
      org.opencontainers.image.version="${VDL_VERSION}"

RUN \
    # Install system dependencies
    apt-get update -qq && \
    apt-get install -yqq --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    # Create group and user
    groupadd -g 65532 nonroot && \
    useradd -u 65532 -g nonroot -m -s /bin/sh nonroot && \
    # Create VDL workdir
    mkdir /vdl && chown nonroot:nonroot /vdl

WORKDIR /vdl

COPY --from=builder --chown=nonroot:nonroot /workspaces/vdl/dist/vdl /usr/local/bin/vdl

USER nonroot

ENTRYPOINT ["/usr/local/bin/vdl"]

CMD ["--help"]
