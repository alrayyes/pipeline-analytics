# Proposal

## Why

An internal tool must not be indexed (alrayyes/dotfiles#810, alrayyes/pipeline-analytics#589). The HTML carries a `robots` meta tag (the `noindex-meta` change), but JSON, static files and error responses have no `<head>`, so the server says it in a header too.

## What Changes

- Every response carries `X-Robots-Tag: noindex, nofollow`, set by a middleware wrapped outside the logger and the session check so rejected requests and health probes carry it as well.
- `robots.txt` stays permissive, so a crawler can fetch a page and read the instruction.
- The README says the tool is internal and where its auth lives.

## Capabilities

### Modified Capabilities

- `site-metadata`: every response asks search engines not to index it.

## Impact

One middleware, its tests, one README paragraph. No behaviour change for visitors.
