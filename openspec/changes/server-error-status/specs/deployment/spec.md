## ADDED Requirements

### Requirement: Errors answer with their real status

The system SHALL answer a request for a path that is neither a file nor a route
of the dashboard with `404` and the not-found page, and a failure to read an
embedded asset with `500` and the error page, never with `200` and the app
shell. The pages SHALL be sent with `Cache-Control: no-cache` (404) or
`no-store` (500), without `noindex`, and without any detail of the failure.
Paths under `/api/` SHALL keep the JSON error shape.

#### Scenario: A path no route matches

- **WHEN** `/no/such/page` is requested
- **THEN** the response is `404` with the not-found page's HTML and
  `Cache-Control: no-cache`

#### Scenario: A route of the dashboard

- **WHEN** `/runs/42` is requested and `/runs/[id]` is one of the dashboard's
  routes
- **THEN** the response is `200` with the app shell

#### Scenario: A file that doesn't exist

- **WHEN** `/_app/immutable/gone.js` is requested
- **THEN** the response is `404`, not the app shell

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
