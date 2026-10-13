# site-metadata Specification

## Purpose

Serves static discovery files about the web app itself — who built it
and what it's built with — separate from the dashboard's own pipeline
data and metrics capabilities.

## Requirements

### Requirement: humans.txt disclosure

The system SHALL serve a `humans.txt` file at the site root, following
the humanstxt.org `/* TEAM */` + `/* SITE */` section convention, naming
the maintainer by public handle only (no email address or other personal
identifying information) and the core stack.

#### Scenario: Visitor requests humans.txt

- **WHEN** a client sends `GET /humans.txt`
- **THEN** the system returns `200` with a `text/plain` body containing
  a `TEAM` section with the maintainer's public handle and a `SITE`
  section naming the stack

#### Scenario: No personal identifying information beyond the handle

- **WHEN** the `humans.txt` content is inspected
- **THEN** it contains no email address, physical location, or other
  personal identifying information beyond the maintainer's public handle

### Requirement: Responses are not indexable

The system SHALL send `X-Robots-Tag: noindex, nofollow` on every response, whatever its status or content type, and SHALL NOT serve a `robots.txt` that disallows crawling, so the instruction can be read.

#### Scenario: A crawler fetches any URL

- **WHEN** a client requests any path, including the app shell, an API endpoint, a health probe, a static file or a path that doesn't exist
- **THEN** the response carries `X-Robots-Tag: noindex, nofollow`

#### Scenario: A request is rejected before any handler runs

- **WHEN** a request is refused for a missing session or an oversized body
- **THEN** the refusal carries the header too

#### Scenario: A crawler fetches robots.txt

- **WHEN** a crawler fetches `/robots.txt`
- **THEN** it is not told to stay away from `/`
