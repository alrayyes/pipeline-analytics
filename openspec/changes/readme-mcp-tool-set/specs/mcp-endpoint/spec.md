# Spec Delta

## ADDED Requirements

### Requirement: The README describes the registered tools without a count

The system SHALL describe the MCP tool set in the README by what the tools do,
including that some of them change data, and SHALL NOT state a tool count or
call the set read-only.

#### Scenario: A tool is added

- **WHEN** a tool is registered for a new operation
- **THEN** the README's description stays true without an edit to a number

#### Scenario: A reader looks for the tools that act

- **WHEN** a reader opens the README's MCP section
- **THEN** it says which tools write and which of those need a session
