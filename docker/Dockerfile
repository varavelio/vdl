# Packages the VDL CLI into a minimal, ready-to-use container image.
# The pre-built binary is fetched directly from GitHub Releases so no compilation
# step is required, keeping the final image as small as possible.
#
# The binary lives at /usr/local/bin/vdl, which is on the default $PATH, so it
# behaves like a native command:
#
#   docker run --rm varavel/vdl:latest
#   docker run --rm varavel/vdl:latest format
#
# Supported platforms
# -------------------
# - linux/amd64
# - linux/arm64

# -- downloader ---------------------------------------------------------------
# Minimal stage that pulls the correct VDL release archive for the target
# architecture from GitHub and extracts the binary.
# -----------------------------------------------------------------------------

FROM alpine:3.23.3 AS downloader

ARG TARGETARCH
ARG VDL_VERSION

RUN apk add --no-cache curl tar

RUN set -e; \
    case "${TARGETARCH}" in \
      amd64) ARCH="amd64" ;; \
      arm64) ARCH="arm64" ;; \
      *)    echo "Unsupported architecture: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    URL="https://github.com/varavelio/vdl/releases/download/v${VDL_VERSION}/vdl_linux_${ARCH}.tar.gz"; \
    echo "Downloading ${URL}"; \
    curl -fsSL "${URL}" -o /tmp/vdl.tar.gz; \
    tar -xzf /tmp/vdl.tar.gz -C /usr/local/bin vdl; \
    chmod +x /usr/local/bin/vdl; \
    /usr/local/bin/vdl version

# -- runtime ------------------------------------------------------------------
# Bare Debian image containing only the VDL binary.
# -----------------------------------------------------------------------------

FROM debian:trixie-slim

ARG VDL_VERSION

LABEL org.opencontainers.image.title="VDL" \
      org.opencontainers.image.authors="Varavel" \
      org.opencontainers.image.description="VDL is an open-source, type-safe, multi-language, and easily extensible schema definition language and code generation toolchain." \
      org.opencontainers.image.url="https://github.com/varavelio/vdl" \
      org.opencontainers.image.documentation="https://github.com/varavelio/vdl" \
      org.opencontainers.image.source="https://github.com/varavelio/vdl" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="Varavel" \
      org.opencontainers.image.version="${VDL_VERSION}"

COPY --from=downloader /usr/local/bin/vdl /usr/local/bin/vdl

ENTRYPOINT ["vdl"]
CMD ["--help"]
