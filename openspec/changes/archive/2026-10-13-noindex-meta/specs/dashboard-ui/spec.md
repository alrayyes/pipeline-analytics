# Spec Delta

## ADDED Requirements

### Requirement: HTML documents are not indexable

Every HTML document the system serves, the app shell and the 404 and 500
pages, SHALL carry `<meta name="robots" content="noindex, nofollow">`, and
`robots.txt` SHALL NOT disallow crawling, so the instruction can be read.

#### Scenario: A crawler fetches the app shell

- **WHEN** a crawler fetches any page of the dashboard
- **THEN** the document's `<head>` asks it not to index the page or follow its
  links

#### Scenario: A crawler fetches robots.txt

- **WHEN** a crawler fetches `/robots.txt`
- **THEN** it is not told to stay away from `/`
