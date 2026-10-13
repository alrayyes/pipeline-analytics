# Spec Delta

## ADDED Requirements

### Requirement: The dependency audit has no ignores

The system SHALL run `bun audit` in CI without `--ignore`, so that a known
advisory in any dependency fails the build.

#### Scenario: An advisory with no fix

- **WHEN** a dependency has an advisory and no patched release
- **THEN** the dependency is overridden or replaced, and the audit stays
  strict
