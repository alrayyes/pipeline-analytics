# Proposal

## Why

Every page is `noindex` on purpose, so Lighthouse's SEO category either fails
on `is-crawlable` or needs that audit skipped (#592). The service is internal;
a search score means nothing for it (#617).

## What Changes

- `lighthouserc.cjs` runs the accessibility, best-practices and performance
  categories only, and drops the `categories:seo` assertion and the
  `is-crawlable` skip.

## Capabilities

### Modified Capabilities

- `ci-pipeline`: the Lighthouse threshold scenario no longer names SEO.

## Impact

- `web/lighthouserc.cjs`.
