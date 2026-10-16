# Spec Delta

## ADDED Requirements

### Requirement: An operation is a WebMCP tool or excluded, never both

The system SHALL fail its tests when an operation in the OpenAPI description has
a WebMCP tool and also carries an `x-agent-excluded` extension.

#### Scenario: A tool for an excluded operation

- **WHEN** an operation has an `x-agent-excluded` reason and a WebMCP tool is
  mapped to it
- **THEN** the test suite fails and names the operation
