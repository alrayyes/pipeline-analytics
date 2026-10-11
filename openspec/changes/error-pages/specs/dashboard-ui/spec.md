# Spec Delta

## ADDED Requirements

### Requirement: Error views in the dashboard's own layout

The system SHALL render a "Page not found" view for a path no route matches and
a "Something went wrong" view for any other error from loading or rendering,
each inside the shared layout, with that text as the `h1` and the document
title, one sentence that does not blame the visitor, and a link to the overview.
The views SHALL NOT show a stack trace, file path, query, trace ID, cluster,
upstream service or version, or an error code as the headline, and SHALL pass an
automated axe-core scan at desktop and 375px width in light and dark.

#### Scenario: An unknown path

- **WHEN** the user opens a path that no route matches
- **THEN** the view reads "Page not found" and offers a link to the overview,
  with the navigation and footer still present

#### Scenario: A failure while loading

- **WHEN** a load function throws
- **THEN** the view reads "Something went wrong", shows nothing about the
  failure, and the detail is written to the console

#### Scenario: Errors the app cannot render

- **WHEN** the error happens outside the app's own layout
- **THEN** a self-contained page with the same wording is shown
