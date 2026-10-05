# Spec Delta

## ADDED Requirements

### Requirement: Local hooks cover every check CI runs

The system SHALL run each check CI runs from a git hook too, on the same file
group, unless `lefthook.yml` names the check as CI-only and says why.

#### Scenario: A Dockerfile change runs hadolint locally

- **WHEN** a push changes the Dockerfile
- **THEN** the hadolint hook runs, as the CI job does

#### Scenario: A GoReleaser change runs the config check locally

- **WHEN** a push changes `.goreleaser.yml`
- **THEN** `goreleaser check` runs

#### Scenario: A go.mod change runs the formatting check locally

- **WHEN** a push changes `go.mod`
- **THEN** the formatting diff CI runs also runs

#### Scenario: A clean machine has no Docker cache directories

- **WHEN** a Go hook runs and the cache directories don't exist
- **THEN** the hook creates them as the user and doesn't fail on permissions

#### Scenario: A passing hook is silent

- **WHEN** every hook passes
- **THEN** lefthook prints nothing, and a failing hook prints its output

#### Scenario: A CI check loses its hook

- **WHEN** a CI check has no hook and isn't named CI-only
- **THEN** the `changes` job fails
