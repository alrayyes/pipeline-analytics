# Spec Delta

## ADDED Requirements

### Requirement: Every operation has a WebMCP tool or a written exclusion

The system SHALL fail its tests when an operation in the OpenAPI description,
of any method, has neither a WebMCP tool nor an `x-agent-excluded` extension
that states why an agent must not reach it.

#### Scenario: An operation without a tool

- **WHEN** an operation is added to the OpenAPI description with no WebMCP
  tool and no `x-agent-excluded` reason
- **THEN** the test suite fails and names the operation

#### Scenario: A tool the test does not know

- **WHEN** a WebMCP tool is registered that the parity test does not map to
  operations
- **THEN** the test suite fails and names the tool
