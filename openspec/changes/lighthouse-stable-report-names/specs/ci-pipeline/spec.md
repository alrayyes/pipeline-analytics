# Spec Delta

## ADDED Requirements

### Requirement: Published Lighthouse reports keep their names between deploys

The system SHALL publish each audited page's Lighthouse report as
`<page>.report.html` and `<page>.report.json`, with no timestamp in the name,
and SHALL write a `manifest.json` whose `htmlPath` and `jsonPath` end in those
names, so that a manifest cached from before a deploy still resolves.

#### Scenario: Two deploys in a row

- **WHEN** the reports are assembled twice from runs made at different times
- **THEN** the published `lighthouse/` directory lists the same file names both
  times

#### Scenario: The manifest names real files

- **WHEN** the reports are assembled
- **THEN** every file the manifest's `htmlPath` and `jsonPath` name exists in
  the `lighthouse/` directory

#### Scenario: A page with more than one run

- **WHEN** the manifest holds several runs of one URL
- **THEN** only its representative run is published, under the page's name

#### Scenario: The root page

- **WHEN** the audited URL is the site root
- **THEN** its report is called `home`
