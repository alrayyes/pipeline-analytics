# syntax=docker/dockerfile:1
# The binary is cross-compiled by goreleaser before this runs -- COPY only,
# never `go build` here (go-releases.md: avoids paying QEMU emulation cost
# per non-native arch in the release matrix).
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

RUN apk add --no-cache ca-certificates=20260909-r0 && \
    addgroup -S -g 10001 pipeline-analytics && \
    adduser -S -u 10001 -G pipeline-analytics pipeline-analytics

COPY pipeline-analytics /usr/local/bin/pipeline-analytics

USER 10001:10001
EXPOSE 8080

# Exec form, not shell form - this image has no curl or wget, only busybox's
# own sh. The binary's own `healthcheck` subcommand exists for exactly this.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/pipeline-analytics", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/pipeline-analytics"]
CMD ["serve"]
