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
