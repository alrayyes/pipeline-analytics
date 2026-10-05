# Proposal

## Why

The telemetry designs show Re-run, Cancel and Quarantine buttons, and the
dashboard has been read-only toward the forges. #346 asked for a decision. The
decision is yes for re-run and cancel; this change is the spec delta the ticket
asks for (which actions, the token scope each needs, how a failure shows). Each
action gets its own ticket.

## Research

- **GitHub** has REST endpoints to re-run and cancel a workflow run, and a
  fine-grained token needs the **Actions** repository permission set to
  **write** for them
  ([GitHub docs: workflow runs](https://docs.github.com/en/rest/actions/workflow-runs)).
  The search that surfaced this showed the permission for the re-run endpoint;
  I did not read the cancel endpoint's page, so that row is unchecked.
- **Forgejo** has the actions only as web routes
  (`POST /{owner}/{repo}/actions/runs/{run}/cancel` and the per-job re-run),
  behind a browser session, and no REST endpoint I could find. This service
  targets 11.0.16. Found by search, not checked against a live instance. Same
  position as the job log (#343): Forgejo answers "unsupported" until a REST API
  exists to call.
- The registration screen asks for **Actions: read-only** today
  (`web/src/routes/repos/+page.svelte`). A write needs more.

## Decisions

- **Two actions, both on a run.** *Re-run* is offered on a concluded run and
  re-runs only the failed jobs when the run failed (the case after a flake),
  the whole run otherwise. *Cancel* is offered on a queued or running run.
- **Quarantine is not a forge action.** It would mark a flaky step inside this
  app, with no forge write, and nobody has said what it should do. It gets its
  own ticket as a question, outside this change.
- **GitHub only for now.** A Forgejo run answers `unsupported`, as its job log
  does.
- **Session-only, never an MCP tool.** These write to someone else's system, so
  an API token or an agent doesn't get them, like passkeys and saved tokens.
  The MCP server stays read-only.
- **The token scope is a user decision, shown honestly.** A stored token with
  read-only Actions gets `forbidden` from GitHub. The UI says the token needs
  Actions write permission to do this and says nothing about re-registering
  anything: replacing the saved token (#462) is how it's fixed. The README and
  the register dialog's guidance say what write needs.
- **Status codes, not a 200 with a flag.** A write that didn't happen shouldn't
  look like success to a script.

| Outcome | Status | `code` |
| --- | --- | --- |
| The forge accepted it | `202` | none |
| Run unknown | `404` | `not_found` |
| Run can't be re-run or cancelled in its state | `409` | `not_actionable` |
| Token can't do it | `403` | `forbidden` |
| Forge has no API for it | `501` | `unsupported` |
| Forge didn't answer | `502` | `unreachable` |

- **Logged.** Each action is logged with the run, the action and the outcome,
  never the token.

## What Changes

- The spec for the two actions and their failures, below. No code in this
  change.
- Follow-up tickets: re-run, cancel, the token guidance, and the Quarantine
  question.

## Capabilities

### New Capabilities

- `forge-actions`: acting on a run on its forge.
