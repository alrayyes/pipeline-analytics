# Proposal

## Why

The Lighthouse job audits `/login` only, because every other page sits behind
a passkey. Performance, best-practices and SEO regressions on the pages people
use go unseen. Tracked in #520.

## What Changes

- The job signs the browser in through a virtual authenticator, then audits
  the signed-in pages beside `/login`: one report per URL.
- The accessibility, best-practices and SEO thresholds apply to every page;
  performance stays warn-only.
- A page that was bounced to `/login` instead of audited fails the job.
- One run per URL, not three: eleven pages at three runs each is too long for
  CI, and the gated categories don't vary between runs.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-pipeline`: adds a requirement for the signed-in Lighthouse audit.

## Impact

- `web/lighthouserc.cjs`, `web/scripts/lighthouse.sh`, new
  `web/scripts/lighthouse-auth.cjs` and `check-lighthouse-reports.ts`.
- `README.md`, `CONTRIBUTING.md` and the job's comment in `ci.yml`.
