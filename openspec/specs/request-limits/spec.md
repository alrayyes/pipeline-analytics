# request-limits Specification

## Purpose
Keeps the server from reading an unbounded request body, so one oversized request cannot exhaust memory: the API, the MCP endpoint and the forge webhook receivers each have a limit and answer an oversized body with a defined error.

## Requirements

### Requirement: Request bodies are bounded

The system SHALL refuse a request body larger than its limit with a 413 and
SHALL NOT read it past that limit: 64 KiB for the API and the MCP endpoint,
2 MiB for the forge webhook receivers.

#### Scenario: A declared oversized body

- **WHEN** a request declares a body longer than the limit
- **THEN** the response is 413 and the body is not read

#### Scenario: An undeclared oversized body

- **WHEN** a request streams a body with no declared length past the limit
- **THEN** the read stops at the limit and the response is 413

#### Scenario: A webhook larger than an API call

- **WHEN** a forge webhook delivery is larger than the API limit and under the
  webhook limit
- **THEN** it is read and judged on its signature, not refused for its size

### Requirement: Request fields are bounded

The system SHALL declare a maximum length for every free-text request field
and SHALL reject a longer value with a 400 before acting on it.

#### Scenario: A token past its limit

- **WHEN** a repo is registered with a token longer than its `maxLength`
- **THEN** the response is 400 and nothing is stored
