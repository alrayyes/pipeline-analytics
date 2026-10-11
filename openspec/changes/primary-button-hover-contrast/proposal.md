# Proposal

## Why

A hovered primary button fails WCAG AA colour contrast. `hover:bg-primary/80`
blends the teal toward the white page, and the white label on it measures about
3.3:1 against the 4.5:1 required. It showed up as an e2e axe flake (#570): the
scan failed whenever the pointer was still over the button after a click.

## What Changes

- The default button darkens on hover (`hover:brightness-90`) instead of
  fading toward the background, so the hover state keeps its contrast.
- The e2e axe scans wait for ending transitions to finish (`scanSettled`), so
  a scan never reads a colour that is part-way through a change.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `dashboard-ui`: adds a requirement that interactive states keep AA contrast.

## Impact

- `web/src/lib/components/ui/button/button.svelte`, `web/tests/e2e/axe.ts` and
  the specs that scanned inline.
