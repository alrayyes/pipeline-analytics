# Spec Delta

## ADDED Requirements

### Requirement: The axe scan survives a cancelled animation

The system SHALL treat an animation cancelled while the e2e axe scan waits for
it as settled, so that a dialog closing mid-scan doesn't fail the build.

#### Scenario: An animation is cancelled mid-wait

- **WHEN** an animation is cancelled while the scan waits for it to finish
- **THEN** the scan carries on instead of failing on the rejected wait
