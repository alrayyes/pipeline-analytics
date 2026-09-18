# release-process Specification

## Purpose

Defines how a merged pull request turns into exactly one changelog and
release-notes entry, independent of which merge strategy GitHub offers.

## Requirements

### Requirement: Squash-only merges

The repository SHALL allow only the squash merge strategy for pull
requests into the default branch.

#### Scenario: Merge button settings are squash-only

- **WHEN** the repo's merge settings are read via the GitHub API
- **THEN** `squash_merge_allowed` is `true` and both
  `merge_commit_allowed` and `allow_rebase_merge` are `false`

### Requirement: One changelog entry per shipped change

Generated changelog and release-notes content SHALL contain exactly one
entry per shipped change.

#### Scenario: A merged feature PR produces one entry

- **WHEN** a pull request is merged and release-please next runs
- **THEN** the resulting `CHANGELOG.md` section and the corresponding
  GitHub release notes each list that change exactly once
