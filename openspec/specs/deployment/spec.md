# Deployment Specification

## Purpose
What the service leaves to its environment, so an operator knows what to provide around it.

## Requirements

### Requirement: Rate limiting is left to the reverse proxy

The system SHALL NOT rate-limit requests itself, and SHALL document in its
deployment guidance which routes are reachable without a session, so an
operator can limit them at the proxy.

#### Scenario: An operator reads the deployment guidance

- **WHEN** an operator sets up a reverse proxy
- **THEN** the README says rate limiting belongs there and lists the routes
  reachable without a session

### Requirement: Assets are cached by content and sent compressed

The system SHALL send the dashboard's content-hashed files with a long
immutable cache lifetime, everything else with revalidation, and SHALL serve a
precompressed copy of a text file to a client that accepts its encoding.

#### Scenario: A hashed file

- **WHEN** a file under `_app/immutable/` is requested
- **THEN** the response has `Cache-Control: public, max-age=31536000, immutable`

#### Scenario: The page and unhashed files

- **WHEN** the page, a client route or an unhashed file is requested
- **THEN** the response has `Cache-Control: no-cache` plus `ETag`, and a
  matching `If-None-Match` gets a 304

#### Scenario: A client that accepts Brotli or gzip

- **WHEN** a text file with a compressed copy is requested with
  `Accept-Encoding: br, gzip`
- **THEN** the response is the Brotli copy, with `Content-Encoding: br`,
  `Vary: Accept-Encoding` and the original file's content type

#### Scenario: A client that accepts neither

- **WHEN** the same file is requested with no `Accept-Encoding`
- **THEN** the response is the file as built, with no `Content-Encoding`

### Requirement: Errors answer with their real status

The system SHALL answer a path that is neither a file nor a route of the
dashboard with status `404`, and a failure to read an embedded asset with
`500`, never with `200`. Error responses SHALL carry `Cache-Control: no-cache`
(404) or `no-store` (500), no `noindex`, and no detail of the failure.

#### Scenario: A path no route matches

- **WHEN** `/no/such/page` is requested
- **THEN** the response is `404` with the app shell and
  `Cache-Control: no-cache`, and the client renders "Page not found"

#### Scenario: A route of the dashboard

- **WHEN** `/runs/42` is requested and `/runs/[id]` is one of the dashboard's
  routes
- **THEN** the response is `200` with the app shell

#### Scenario: A file that doesn't exist

- **WHEN** `/_app/immutable/gone.js` is requested
- **THEN** the response is a plain `404`, not the app shell

#### Scenario: The shell can't be read for a 404

- **WHEN** a path no route matches is requested and the app shell is unreadable
- **THEN** the response is `404` with the static `404.html`

#### Scenario: An unknown API path

- **WHEN** `/api/no-such-endpoint` is requested
- **THEN** the response is the JSON `404` the API already sends, not HTML

#### Scenario: An embedded asset can't be read

- **WHEN** the server can't read the file it needs to answer a request
- **THEN** the response is `500` with the error page and
  `Cache-Control: no-store`, the body names no path or error text, and the
  detail is logged

#### Scenario: The route list comes from the frontend's routes

- **WHEN** a page is added under the frontend's `src/routes`
- **THEN** the next build's `routes.json` lists it, with no second list to
  update
