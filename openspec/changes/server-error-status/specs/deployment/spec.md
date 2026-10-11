## ADDED Requirements

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
