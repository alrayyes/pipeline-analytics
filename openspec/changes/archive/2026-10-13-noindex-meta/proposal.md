# Proposal

## Why

An internal tool must not be indexed (alrayyes/dotfiles#810). The server sends
`X-Robots-Tag: noindex, nofollow` (backend half of #589); the HTML carries the
same instruction so a crawler that ignores headers still reads it.

## What Changes

- `<meta name="robots" content="noindex, nofollow">` in `src/app.html`,
  `src/error.html` and `static/404.html`, `static/500.html`.
- `static/robots.txt` stays permissive: a `Disallow: /` would stop a crawler
  fetching the page, so it would never see the noindex.
- A unit test fails if any of those documents loses the tag or `robots.txt`
  starts disallowing everything.

## Capabilities

### Modified Capabilities

- `dashboard-ui`: every HTML document asks search engines not to index it.

## Impact

Four HTML files and one test. No behaviour change for visitors.
