# Proposal

## Why

The pipeline runs `bun audit` and `govulncheck` on every push and pull request.
An advisory published overnight turns every open pull request red, including
ones that touch no dependency, and a known advisory with no fix pushes people
toward an ignore list that hides it from the only place that tracks it (#657).

## Research

- The shared action `alrayyes/advisory-scan` has a gate mode (audit the base and
  the head, fail on the IDs the head adds) and a scan mode (keep one issue per
  title for the default branch's findings). It takes no ignore list. Its README
  documents `bun audit` and govulncheck as supported commands.
- Two lockfiles, so two scan issues with different titles. The action finds its
  issue by title, and one title shared by two jobs would have them overwrite
  each other.

## What changes

- Both audit steps in `ci.yml` become the action in gate mode, pinned to a SHA.
- A new `advisory-scan.yml` runs scan mode daily, on lockfile pushes and on
  dispatch, one job for Go and one for the web app.
- `CONTRIBUTING.md` says how the two fit together.

## Impact

- Modified capability: `ci-pipeline`. The "no ignores" requirement keeps its
  intent and now covers both audits and the scan.
