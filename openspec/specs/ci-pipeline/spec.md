# CI pipeline Specification

## Purpose
Keeps the pipeline and the local git hooks saying the same thing: what runs, on which files, and what fails the build.

## Requirements

### Requirement: Local hooks cover every check CI runs

The system SHALL run each check CI runs from a git hook too, on the same file
group, unless `lefthook.yml` names the check as CI-only and says why.

#### Scenario: A Dockerfile change runs `hadolint` locally

- **WHEN** a push changes the Dockerfile
- **THEN** the `hadolint` hook runs, as the CI job does

#### Scenario: A `GoReleaser` change runs the config check locally

- **WHEN** a push changes `.goreleaser.yml`
- **THEN** `goreleaser check` runs

#### Scenario: A `go.mod` change runs the formatting check locally

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

### Requirement: Test results reach Codecov for both suites

The system SHALL write JUnit results from the Go and frontend tests and upload
them to Codecov as test results, including when a test failed, except on
Dependabot runs, which have no upload token.

#### Scenario: A test fails

- **WHEN** a Go or frontend test fails
- **THEN** its JUnit results are still uploaded

#### Scenario: A Dependabot run

- **WHEN** Dependabot opens a pull request
- **THEN** no test results are uploaded and the job doesn't fail on a missing
  token

#### Scenario: The upload input is misspelled

- **WHEN** the workflow says `report-type` instead of `report_type`
- **THEN** the `changes` job fails

### Requirement: A surviving frontend mutant fails the build

The system SHALL fail the frontend mutation job, and the matching pre-push
hook, when any mutant in the mutated modules survives, unless the mutant is
equivalent and the code says why.

#### Scenario: A test stops pinning a behavior

- **WHEN** a change leaves a mutant in a mutated module surviving
- **THEN** the mutation job fails and the report names the mutant

#### Scenario: An equivalent mutant

- **WHEN** a mutant can't change behavior
- **THEN** a Stryker disable comment beside it says why, and it doesn't fail
  the job
