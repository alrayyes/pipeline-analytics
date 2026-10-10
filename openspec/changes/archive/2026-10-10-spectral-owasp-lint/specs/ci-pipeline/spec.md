# Spec Delta

## ADDED Requirements

### Requirement: The API description is linted for security

The system SHALL lint the OpenAPI description against the OWASP API Security
ruleset whenever the spec or its tooling changes, SHALL fail the run on an
error-severity finding, and SHALL state in the ruleset's configuration why each
disabled rule is disabled.

#### Scenario: A spec change that breaks a security rule

- **WHEN** a pull request changes the spec so it breaks an enabled
  error-severity rule
- **THEN** the `spectral (owasp)` job fails and names the rule

#### Scenario: A rule is turned off

- **WHEN** a rule is disabled in `.spectral.yaml`
- **THEN** a comment beside it says why, and names the ticket that revisits it
  where one exists
