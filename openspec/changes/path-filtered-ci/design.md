# Design

## Decisions

**Job-level `if:` behind a `changes` job, not workflow-level `paths:`.**
`ci.yml` is one workflow, and main's ruleset requires nine of its checks.
GitHub reports a job skipped by a conditional as a success, where a
workflow skipped by a path filter stays pending and blocks a pull request
that requires it
([GitHub docs](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/troubleshooting-required-status-checks)).

**Groups follow what a job reads, not its name.** Several are wider than they
look, and the script's test pins each:

- The Go tests also cover `openapi/**` (the MCP parity test parses the spec)
  and the embedded frontend directory.
- e2e and Lighthouse build and run the real binary, so they cover both the
  frontend and Go.
- The frontend build reads `CHANGELOG.md` for the release page (#368), so a
  changelog change runs the frontend jobs.
- Prettier checks `**/*.{md,yml,yaml}`, so any YAML change, including the
  OpenAPI spec and workflow files, runs the formatting check.

**Pipeline edits run everything.** A change to `ci.yml` or the filter script
turns every group on, so a mistake in the filter can't hide behind the filter.

**Unknown base runs everything.** A new branch or a force push that dropped
the base can't say what changed, and filtering on a guess would skip checks.

**The filter tests itself in the always-on job.** The `changes` job runs
`changes.test.sh` before computing anything, so a bad path fails CI instead
of silently skipping a check.

**Never filtered.** Release, PR-title lint, auto-merge and SDK-notify live in
other workflows and are untouched. The `changes` job always runs, so every
pull request has a status.

## Risks / Trade-offs

- **A job's real inputs can be wider than its group.** The mitigation is the
  case-per-path test, and the rule of adding a case first. A job that starts
  reading a new kind of file needs its group widened.
- **Skipped jobs show as skipped, not as run.** That is the point, and it
  also means a green pull request proves less about files it didn't touch.
- **The Pages deploy no longer reruns on every push to main**, only when the
  spec or `docs/api/` changed, which is all it publishes.
- **`lefthook.yml` uses its own globs.** The rule wants hook and workflow
  paths identical; that alignment is a separate follow-up.
