# Spec Delta

## REMOVED Requirements

### Requirement: Every read endpoint has a tool

**Reason**: Replaced by "Every operation has a tool or a written exclusion",
which covers every method, not only GET. Its scenarios (a run an agent finds
can be opened, failure insights over MCP) are per-tool behaviour the tool
tests already pin.

**Migration**: None. The read tools stay as they are.

## ADDED Requirements

### Requirement: Every operation has a tool or a written exclusion

The system SHALL fail its tests when an operation in the OpenAPI description,
of any method, has neither an MCP tool nor an `x-agent-excluded` extension
that states why an agent must not reach it.

#### Scenario: A write operation without a tool

- **WHEN** a POST, PUT, PATCH or DELETE operation is added to the OpenAPI
  description with no tool and no `x-agent-excluded` reason
- **THEN** the test suite fails and names the operation

#### Scenario: An exclusion with no reason

- **WHEN** an operation carries an empty `x-agent-excluded`
- **THEN** the test suite fails and names the operation
