# Spec Delta

## ADDED Requirements

### Requirement: Every WebMCP tool declares an input schema generated from the spec

The system SHALL register each WebMCP tool with a JSON Schema input derived
from the path and query parameters of the OpenAPI operation it wraps, and
SHALL fail its tests when a tool's schema or the committed generated module no
longer matches the spec.

#### Scenario: A spec parameter changes

- **WHEN** a parameter of a wrapped operation changes in the OpenAPI
  description and the schema is not regenerated
- **THEN** the test suite fails

#### Scenario: A tool pages by default where the API does not

- **WHEN** an agent reads `list_runs` or `list_pipelines`
- **THEN** the description and the schema state the default of 20 items from
  offset 0
