# Spec Delta

## ADDED Requirements

### Requirement: The settings response carries the defaults

The system SHALL return, with every settings response, the documented default
of each setting, whatever the account has stored.

#### Scenario: A stored value differs from its default

- **WHEN** an account has set the Pipelines health filter to `all`
- **THEN** the response's `pipelinesHealthFilter` is `all` and
  `defaults.pipelinesHealthFilter` is `unhealthy`
