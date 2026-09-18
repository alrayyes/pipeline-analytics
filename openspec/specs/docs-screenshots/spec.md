# docs-screenshots Specification

## Purpose

Keeps the README's screenshots capturable against a real running
server, so the capture script doesn't silently fall behind when the
UI it drives changes shape.

## Requirements

### Requirement: Capture script matches the current UI

`web/scripts/capture-screenshots.mjs` SHALL locate every control it
drives by a selector that resolves to exactly one element in the
current UI.

#### Scenario: Health filter click is unambiguous

- **WHEN** the script switches the Pipelines overview to show healthy
  and unhealthy pipelines
- **THEN** it scopes the click to the health-status radiogroup, not a
  bare selector that also matches the forge filter's own "All" radio

#### Scenario: Full capture run succeeds

- **WHEN** `web/scripts/capture-screenshots.sh` is run against a fresh
  server
- **THEN** it completes and prints "Screenshots captured." with exit
  code `0`
