# Spec Delta

## ADDED Requirements

### Requirement: Assets are cached by content and sent compressed

The system SHALL send the dashboard's content-hashed files with a long
immutable cache lifetime, everything else with revalidation, and SHALL serve a
precompressed copy of a text file to a client that accepts its encoding.

#### Scenario: A hashed file

- **WHEN** a file under `_app/immutable/` is requested
- **THEN** the response has `Cache-Control: public, max-age=31536000, immutable`

#### Scenario: The page and unhashed files

- **WHEN** the page, a client route or an unhashed file is requested
- **THEN** the response has `Cache-Control: no-cache` and an `ETag`, and a
  matching `If-None-Match` gets a 304

#### Scenario: A client that accepts Brotli or gzip

- **WHEN** a text file with a compressed copy is requested with
  `Accept-Encoding: br, gzip`
- **THEN** the response is the Brotli copy, with `Content-Encoding: br`,
  `Vary: Accept-Encoding` and the original file's content type

#### Scenario: A client that accepts neither

- **WHEN** the same file is requested with no `Accept-Encoding`
- **THEN** the response is the file as built, with no `Content-Encoding`
