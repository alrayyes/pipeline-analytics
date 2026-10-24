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
The audit covers accessibility, best practices and performance, and not SEO,
because the service is internal, and every page is `noindex`.

#### Scenario: Every page has a report

- **WHEN** the lighthouse job finishes
- **THEN** its report artifact has one report for `/login` and one for each
  signed-in page

#### Scenario: A signed-in page below a threshold

- **WHEN** a signed-in page scores below the accessibility or best-practices
  threshold
- **THEN** the job fails, and a performance score below its threshold only
  warns

#### Scenario: A page that was not reached

- **WHEN** a page's report ended on a different URL than the one requested
- **THEN** the job fails

#### Scenario: No SEO score

- **WHEN** the lighthouse job finishes
- **THEN** its reports hold no SEO category

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

### Requirement: The dependency audit has no ignores

The system SHALL run `bun audit` in CI without `--ignore`, so that a known
advisory in any dependency fails the build.

#### Scenario: An advisory with no fix

- **WHEN** a dependency has an advisory and no patched release
- **THEN** the dependency is overridden or replaced, and the audit stays
  strict

### Requirement: The axe scan survives a cancelled animation

The system SHALL treat an animation cancelled while the e2e axe scan waits for
it as settled, so that a dialog closing mid-scan doesn't fail the build.

#### Scenario: An animation is cancelled mid-wait

- **WHEN** an animation is cancelled while the scan waits for it to finish
- **THEN** the scan carries on instead of failing on the rejected wait

### Requirement: A Codecov upload error doesn't fail the build

The system SHALL report a failed Codecov upload as a warning and SHALL NOT fail
the job over it, so that an outage on Codecov's side doesn't turn main red. The
test and coverage steps keep failing the job when the code is wrong.

#### Scenario: Codecov refuses an upload

- **WHEN** Codecov answers an upload with a 403 or a 5xx
- **THEN** the job still passes, and the log shows the error

#### Scenario: A test fails while Codecov is down

- **WHEN** a test fails
- **THEN** the job fails at the test step, and the test results still upload

### Requirement: The API docs deployment waits for its artifact

The system SHALL wait until the run's artifact list names the `github-pages`
artifact before it deploys. A list that trails the upload then doesn't fail the
deployment.

#### Scenario: The list trails the upload

- **WHEN** the deployment job starts while the artifact list omits
  `github-pages`
- **THEN** the job looks again for up to a minute, then deploys once it shows

#### Scenario: The artifact never shows

- **WHEN** the list still doesn't name `github-pages` after a minute
- **THEN** the job fails and its log says the artifact is missing

### Requirement: CI tools run on the bun version CI pins

The system SHALL keep the Stryker runner for bun at a release that runs on the
bun version the CI jobs pin, and SHALL stop Dependabot from proposing a release
it cannot run there: one that needs a newer bun, or that fails its dry run on
the CI runner.

#### Scenario: A web change runs the mutation job

- **WHEN** a pull request changes a file under `web/`
- **THEN** the mutation job starts Stryker and ends on its mutation score, not
  on a bun version check

#### Scenario: A runner release CI cannot run

- **WHEN** a release of `@hughescr/stryker-bun-runner` from 1.5.0 is published,
  which times out the dry run in CI or needs bun 1.4 or newer
- **THEN** Dependabot does not propose it while CI pins bun 1.3

### Requirement: The screenshot refresh pull request merges itself on green

The system SHALL open the README screenshot refresh pull request so that its
checks run, and SHALL arm auto-merge on it, so it merges once those checks
pass and not before. The pull request SHALL change only `docs/` and
`README.md`.

#### Scenario: A release changes the UI

- **WHEN** a release changes what the screenshots show
- **THEN** a pull request holding the new screenshots is opened, its checks
  run, and it merges once they pass

#### Scenario: A release does not change the UI

- **WHEN** the screenshots and the README section regenerate unchanged
- **THEN** no pull request is opened

#### Scenario: A pull request is already open

- **WHEN** an earlier refresh pull request is still open at the next release
- **THEN** its branch is replaced with the new screenshots and auto-merge is
  armed on it again

### Requirement: One end-to-end journey scans both color schemes

The system SHALL scan a page with axe in light and in dark from one end-to-end
journey, rather than repeating the sign-in and page load once per scheme.

#### Scenario: Both schemes are scanned

- **WHEN** the end-to-end suite runs an accessibility journey for a page
- **THEN** the page is scanned in light mode and again in dark mode, and a
  violation in either fails the test naming the scheme

#### Scenario: The dark scan is really dark

- **WHEN** the journey scans in dark mode
- **THEN** the page's `dark` class is set before the scan, so the scan cannot
  silently repeat the light one

### Requirement: The CI Playwright image matches the lockfile

The system SHALL run the Playwright jobs in the `ci-playwright` image pinned by
exact version and digest, SHALL keep that version equal to the
`@playwright/test` version in `web/bun.lock`, and SHALL fail a check when they
differ.

#### Scenario: The image is older than the lockfile

- **WHEN** `web/bun.lock` locks `@playwright/test` 1.64.1 and a workflow uses
  the `ci-playwright:1.63.0` image
- **THEN** the `changes` job and the pre-push hook fail, naming both versions

#### Scenario: The image is not pinned by digest

- **WHEN** a workflow names the image by tag alone, or two workflows use
  different digests
- **THEN** the check fails

#### Scenario: A job installs the browsers itself

- **WHEN** a workflow runs `playwright install`, or the container lacks
  `--ipc=host`
- **THEN** the check fails
