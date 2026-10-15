# Spec Delta

## ADDED Requirements

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
