# Spec Delta

## Purpose

Exposes the dashboard's already-fetched pipeline overview, pipeline
detail/trend, and repo usage data as WebMCP tools, so an in-browser
agent sharing the page with its user can read the same data the
dashboard renders instead of parsing the rendered page or a human
relaying it.

## ADDED Requirements

### Requirement: WebMCP tools registered when the API is available

The system SHALL register its WebMCP tools via
`document.modelContext.registerTool()` only when `document.modelContext`
is present, and SHALL NOT error or degrade the dashboard when it is
absent.

#### Scenario: A WebMCP-capable browser loads the dashboard

- **WHEN** the dashboard loads in a browser exposing
  `document.modelContext`
- **THEN** the system registers its tools and they're discoverable by
  that browser's agent

#### Scenario: A browser without WebMCP support loads the dashboard

- **WHEN** the dashboard loads in a browser with no
  `document.modelContext`
- **THEN** the dashboard renders and functions exactly as it does
  today, with no error and no attempt to register a tool

### Requirement: Tools cover the overview, pipeline detail, and usage views

The system SHALL expose, as WebMCP tools, at least: the pipeline
overview list, a single pipeline's health status and trend data, and a
tracked repository's usage.

#### Scenario: Reading the pipeline overview

- **WHEN** an in-browser agent calls the tool covering the overview
  view
- **THEN** the result includes each tracked pipeline's ID, name, and
  current health status, matching what the overview page itself
  displays

#### Scenario: Reading a pipeline's detail

- **WHEN** an in-browser agent calls the tool covering a single
  pipeline's detail
- **THEN** the result includes its duration trend and failure-rate
  trend, matching what the pipeline detail page itself displays

### Requirement: No write tool

The system SHALL NOT expose any WebMCP tool that creates, modifies, or
deletes data.

#### Scenario: Tool list contains no mutating tool

- **WHEN** an in-browser agent lists the dashboard's registered tools
- **THEN** none of them registers a repository, untracks a repository,
  or otherwise changes stored state

### Requirement: Tool data comes from the same fetch path the page uses

The system SHALL derive a WebMCP tool's result from the same function
the corresponding page uses to fetch and render that data, rather than
a separate implementation of the same request.

#### Scenario: A page's fetch logic changes

- **WHEN** the function a page uses to fetch its data changes (a new
  query parameter, a different endpoint)
- **THEN** the WebMCP tool covering that same data reflects the change
  automatically, without a second edit
