# Proposal

## Why

A `null` run, job or step in a Forgejo response decodes to a nil pointer, and
the client dereferenced it, so one malformed reply panicked the service.
Found by fuzzing (#425). Tracked in #446.

## What Changes

- `ListRecentRuns` skips a null run, job or step instead of dereferencing it.

## Capabilities

### Modified Capabilities

- `forge-ingestion`: a malformed forge response can't crash ingestion.
