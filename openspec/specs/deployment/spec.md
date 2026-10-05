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
