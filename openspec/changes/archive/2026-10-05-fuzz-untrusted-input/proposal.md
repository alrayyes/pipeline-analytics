# Proposal

## Why

Nothing fuzzes the code that decodes data we don't control, and the first
fuzz run found a real panic (#446). `rules/go-test.md` asks for native
fuzzing there and a longer pass in CI. Tracked in #425. Stacked on #447,
whose fix the Forgejo seeds need.

## What Changes

- Fuzz tests for the webhook signature check and event decoding, the GitHub
  and Forgejo response decoders, the webhook endpoint and the MCP endpoint,
  seeded with real payloads.
- A `Fuzz` workflow runs each target for a bounded time weekly and on demand,
  and uploads a crashing input.

## Capabilities

### Modified Capabilities

- `forge-ingestion`: decoding of external data is fuzzed.
