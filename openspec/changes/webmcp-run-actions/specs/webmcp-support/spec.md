# Spec Delta

## REMOVED Requirements

### Requirement: No write tool

**Reason**: An in-browser agent needs the writes the dashboard offers
(`rules/mcp.md`). The rule that mattered, that an agent must not change stored
state or act on a forge on its own, is kept and stated below.

**Migration**: None. The read tools are unchanged.

## ADDED Requirements

### Requirement: Write tools act only after a person confirms

The system SHALL let a WebMCP tool that writes, to the forge or to stored
state, act only after the person using the page confirms that call, SHALL send
nothing when they decline or when the page has no way to ask, and SHALL send
the same request as the matching page action so the endpoint's own checks
apply.

#### Scenario: Re-running a run

- **WHEN** an in-browser agent calls `rerun_run` and the person confirms
- **THEN** the same request as the Runs page's re-run button is sent and the
  result reports it as accepted

#### Scenario: The person declines

- **WHEN** an in-browser agent calls `cancel_run` and the person declines
- **THEN** no request is sent and the call fails saying the person declined

#### Scenario: The page cannot ask

- **WHEN** neither `requestUserInteraction` nor `confirm` is available
- **THEN** the call is treated as declined and no request is sent

#### Scenario: The endpoint refuses

- **WHEN** the person confirms and the endpoint refuses, for example because
  the run has not finished
- **THEN** the call fails with the sentence the Runs page shows for that refusal
