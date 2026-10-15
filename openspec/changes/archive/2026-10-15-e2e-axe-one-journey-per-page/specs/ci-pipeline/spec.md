# Spec Delta

## ADDED Requirements

### Requirement: One end-to-end journey scans both colour schemes

The system SHALL scan a page with axe in light and in dark from one end-to-end
journey, rather than repeating the sign-in and page load once per scheme.

#### Scenario: Both schemes are scanned

- **WHEN** the end-to-end suite runs an accessibility journey for a page
- **THEN** the page is scanned in light mode and again in dark mode, and a
  violation in either fails the test naming the scheme

#### Scenario: The dark scan is really dark

- **WHEN** the journey scans in dark mode
- **THEN** the page's `dark` class is set before the scan, so the scan cannot
  silently repeat the light one
