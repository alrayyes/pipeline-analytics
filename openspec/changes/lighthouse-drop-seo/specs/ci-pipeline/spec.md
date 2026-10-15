# Spec Delta

## MODIFIED Requirements

### Requirement: Lighthouse audits the signed-in pages

The system SHALL audit the login page and the pages behind the passkey with
Lighthouse in CI, as the page itself and not as a redirect to the login page.
The audit covers accessibility, best practices and performance, and not SEO,
because the service is internal and every page is `noindex`.

#### Scenario: Every page has a report

- **WHEN** the lighthouse job finishes
- **THEN** its report artifact has one report for `/login` and one for each
  signed-in page

#### Scenario: A signed-in page below a threshold

- **WHEN** a signed-in page scores below the accessibility or best-practices
  threshold
- **THEN** the job fails, and a performance score below its threshold only
  warns

#### Scenario: A page that was not reached

- **WHEN** a page's report ended on a different URL than the one requested
- **THEN** the job fails

#### Scenario: No SEO score

- **WHEN** the lighthouse job finishes
- **THEN** its reports hold no SEO category
