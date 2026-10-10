# Spec Delta

## MODIFIED Requirements

### Requirement: Repo/pipeline overview

The system SHALL present, as the first page shown after login, every
tracked repository with an aggregate health status and the names of
its currently unhealthy pipelines, filterable to all/healthy/unhealthy
and defaulting to unhealthy.

#### Scenario: Unhealthy pipeline is visible at a glance

- **WHEN** the user opens the overview
- **THEN** the system displays every unhealthy pipeline's name against
  its repo, without requiring the user to open each repo individually

#### Scenario: Overview defaults to unhealthy

- **WHEN** the user opens the overview without changing the health
  filter
- **THEN** the system shows only repos with at least one unhealthy
  pipeline

## ADDED Requirements

### Requirement: Repo-scoped pipeline list

The system SHALL present, for a selected repo, a list of its
pipelines with each pipeline's current health status, filterable to
all/healthy/unhealthy and sortable, matching the overview's own
filter and sort options.

#### Scenario: Selecting a repo navigates to its pipelines

- **WHEN** the user selects a repo from the overview
- **THEN** the system navigates to a page listing only that repo's
  pipelines with their health status
