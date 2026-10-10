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

### Requirement: Local hooks run on the files CI runs on

The system SHALL decide whether a git hook runs from the same file groups as
the CI job that runs the same tool, and SHALL fail a hook that names a group
the filter doesn't define.

#### Scenario: A spec-only change runs the Go tests locally

- **WHEN** a push changes only `openapi/openapi.yaml`
- **THEN** the Go test hook runs, as the CI job does

#### Scenario: A docs-only change skips code hooks

- **WHEN** a push changes only documentation
- **THEN** no Go or frontend hook runs

#### Scenario: A hook names a group that doesn't exist

- **WHEN** a hook asks for a group `changes.sh` doesn't define
- **THEN** the hook fails instead of skipping

### Requirement: Lighthouse audits the signed-in pages

The system SHALL audit the login page and the pages behind the passkey with
Lighthouse in CI, as the page itself and not as a redirect to the login page.

#### Scenario: Every page has a report

- **WHEN** the lighthouse job finishes
- **THEN** its report artifact has one report for `/login` and one for each
  signed-in page

#### Scenario: A signed-in page below a threshold

- **WHEN** a signed-in page scores below the accessibility, best-practices or
  SEO threshold
- **THEN** the job fails, and a performance score below its threshold only
  warns

#### Scenario: A page that was not reached

- **WHEN** a page's report ended on a different URL than the one requested
- **THEN** the job fails

### Requirement: Checks run only when the files they cover change

The system SHALL run each CI check only when a file it covers, the tool's own
configuration or lockfile, or the pipeline definition itself changed, and
SHALL skip it otherwise without failing the pull request.

#### Scenario: A docs-only change skips code checks

- **WHEN** a pull request changes only documentation
- **THEN** the Go, frontend, end-to-end, Lighthouse, Docker and spec checks
  are skipped and the prose and markdown checks run

#### Scenario: A code-only change skips prose checks

- **WHEN** a pull request changes only Go source
- **THEN** the prose and markdown checks are skipped and the Go checks, the
  end-to-end check and the Docker build run

#### Scenario: A skipped required check does not block the merge

- **WHEN** a required check is skipped because nothing it covers changed
- **THEN** the pull request is still mergeable

#### Scenario: Editing the pipeline runs every check

- **WHEN** a pull request changes the CI workflow or its filter script
- **THEN** every check runs

#### Scenario: An unknown base runs every check

- **WHEN** the changed files can't be determined, as on a new branch
- **THEN** every check runs

### Requirement: Some workflows are never filtered

The system SHALL run the release, pull-request title lint and merge
automation workflows regardless of which files changed, and SHALL keep at
least one always-running job, so every pull request has a status.

#### Scenario: A docs-only push to main still releases

- **WHEN** a push to main changes only documentation
- **THEN** the release workflow still runs

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

### Requirement: Lighthouse gates render-blocking requests and request chains

The system SHALL fail the lighthouse job when a page has render-blocking
requests or a critical request chain, on `/login` and on each signed-in page.

#### Scenario: Stylesheet requested separately from the page

- **WHEN** the dashboard's stylesheet is served as its own render-blocking
  request
- **THEN** `render-blocking-insight` scores below 0.9 and the job fails

#### Scenario: A resource discovered behind another

- **WHEN** a critical resource such as the web font is only requested after
  the stylesheet has loaded
- **THEN** `network-dependency-tree-insight` scores below 0.9 and the job fails
