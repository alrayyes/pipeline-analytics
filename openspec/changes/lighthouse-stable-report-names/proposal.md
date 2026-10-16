# Proposal

## Why

Lighthouse CI names every report file for its URL and the minute it ran, so each
deploy renames all of them. GitHub Pages serves `manifest.json` and the
`lighthouse/` index with a ten-minute cache behind a CDN, and the CDN ignores a
query string. For up to ten minutes after a deploy the report page can read a
cached manifest that names files the deploy removed, and its links lead to 404
pages (#692). hush-hush and movie-planner-web already publish stable names.

## What changes

- `assemble-reports.sh` publishes one pair per audited page as
  `<page>.report.html` and `<page>.report.json`, named for the path of the URL, with
  the root called `home`.
- It writes a `manifest.json` whose `htmlPath` and `jsonPath` are those names,
  keeping each entry's URL and summary.
- Only the representative run of a page is published. The config runs each page
  once today, so nothing is dropped.
- The deploy job fetches a published report from the live site, as it does for
  the coverage file.

## Impact

- Modified capability: `ci-pipeline`, a new requirement beside the Lighthouse
  ones.
