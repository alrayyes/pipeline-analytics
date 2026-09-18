# Design

## Context

release-please builds `CHANGELOG.md` and each GitHub release's notes
from conventional-commit messages reachable in `main`'s history since
the last release tag. When GitHub merges a PR with "create a merge
commit," it adds a synthetic commit whose body repeats the PR title —
which is itself a valid conventional-commit line — on top of the
original branch commit(s), which are also still reachable as parents.
release-please parses both. Rebase merge would have the same failure
mode (each rebased commit re-lands individually with its own message,
still one entry per commit — but a squash PR title left un-squashed
across several commits still risks multiple entries if more than one
carries a `feat`/`fix` header).

## Goals / Non-Goals

**Goals:**

- Guarantee one entry per merged PR going forward.
- Correct the two already-published releases without rewriting `main`'s
  history or the existing git tags.

**Non-Goals:**

- Reworking release-please's own configuration or commit parsing.
- Rewriting git history (`git filter-branch`/rebase over already-tagged
  commits) to remove the duplicate commits — the tags `v0.29.0` and
  `v0.28.1` are already published and pulled by anyone who's fetched
  them; rewriting history under a published tag is worse than leaving
  the tag's tree alone and just fixing the generated text.

## Decisions

- **Fix the merge strategy, not release-please's parser.** Squash-only
  is already how release-please's own release PRs get merged (#170,
  #169), so it's already the repo's real convention — the other two
  merge methods being enabled was the gap, not release-please missing a
  dedup step. Teaching release-please to dedupe conventional-commit
  bodies would paper over the actual inconsistency and could hide a
  real second entry in a future edge case.
- **Edit the release notes and CHANGELOG.md text directly, not the git
  history.** `gh release edit` changes only the stored notes for a tag;
  `CHANGELOG.md` is a normal file edit on top of `main`. Neither touches
  the tag's commit or tree, so nothing that already fetched `v0.29.0`
  or `v0.28.1` needs to re-fetch.

## Risks / Trade-offs

- Restricting merge methods changes what any future collaborator (or
  Renovate/Dependabot) can pick at merge time → mitigation: squash is
  already the default recommended method and what every other merged
  PR in this repo's history used; no workflow currently depends on
  merge or rebase merges.
