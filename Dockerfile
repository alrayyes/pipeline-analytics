# syntax=docker/dockerfile:1
# The binary is cross-compiled by goreleaser before this runs -- COPY only,
# never `go build` here (go-releases.md: avoids paying QEMU emulation cost
# per non-native arch in the release matrix).
FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce

RUN apk add --no-cache ca-certificates=20260909-r0 && \
    addgroup -S -g 10001 pipeline-analytics && \
    adduser -S -u 10001 -G pipeline-analytics pipeline-analytics

COPY pipeline-analytics /usr/local/bin/pipeline-analytics

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/pipeline-analytics"]
CMD ["serve"]
