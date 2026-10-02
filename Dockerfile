# syntax=docker/dockerfile:1
# The binary is cross-compiled by goreleaser before this runs -- COPY only,
# never `go build` here (go-releases.md: avoids paying QEMU emulation cost
# per non-native arch in the release matrix).
FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

# /data is created and owned by the non-root user here so a fresh named volume
# mounted on it inherits that ownership: Docker copies the image directory's
# owner onto an empty volume, and without the directory the volume comes up
# root-owned and the server can't open its database.
RUN apk add --no-cache ca-certificates=20260909-r0 && \
    addgroup -S -g 10001 pipeline-analytics && \
    adduser -S -u 10001 -G pipeline-analytics pipeline-analytics && \
    mkdir /data && \
    chown 10001:10001 /data

COPY pipeline-analytics /usr/local/bin/pipeline-analytics

USER 10001:10001
EXPOSE 8080

# Readiness, not just liveness: the `healthcheck` subcommand asks /readyz,
# which reads the database, so a container whose SQLite file is unreadable
# reports unhealthy instead of healthy because the process still answers.
# Exec form, not shell form - this image has no curl or wget, only busybox's
# own sh, and the subcommand exists for exactly this.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/pipeline-analytics", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/pipeline-analytics"]
CMD ["serve"]
