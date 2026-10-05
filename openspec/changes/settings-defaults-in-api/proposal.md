# Proposal

## Why

The dashboard kept its own copy of the settings defaults, so a default changed
on the server would leave "Reset filters" wrong. `rules/frontend.md` calls a
constant that mirrors backend config a smell. Tracked in #452 (backend half;
the front end drops its copies in a second pull request).

## What Changes

- `GET` and `PATCH /api/settings` return `defaults` beside the settings in
  force: the server's documented default for each setting.
- The `Settings` schema becomes `SettingsValues` plus `defaults`.
- Purely additive: existing clients ignore the new field.

## Capabilities

### Modified Capabilities

- `account-settings`: the response says what the defaults are.
