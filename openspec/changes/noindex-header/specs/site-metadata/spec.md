# Spec Delta

## ADDED Requirements

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
