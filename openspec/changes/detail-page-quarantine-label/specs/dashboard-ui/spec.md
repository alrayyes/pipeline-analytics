# Spec Delta

## ADDED Requirements

### Requirement: Quarantined steps are labelled on the pipeline detail page

The system SHALL show a "Quarantined" label, in words, beside the flaky marker
of a quarantined step in the pipeline detail page's steps table, and SHALL
leave the label off a step that is not quarantined. The page SHALL pass an
automated axe-core scan with a quarantined step on screen, at phone width too.

#### Scenario: A quarantined flaky step

- **WHEN** the user opens a pipeline whose flaky step is quarantined
- **THEN** that step's row shows "flaky" and "Quarantined"

#### Scenario: A flaky step that is not quarantined

- **WHEN** the same pipeline has another flaky step that is not quarantined
- **THEN** that step's row shows "flaky" and no "Quarantined" label
