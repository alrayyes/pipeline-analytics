# Tasks

## 1. Contract

- [x] 1.1 `Outcome` schema in `openapi.yaml`; `outcome` required on `RunStep`
      and `RunSummary`; Redocly lint

## 2. Backend

- [ ] 2.1 Failing tests: the mapping table, and `outcome` on `GET /api/runs`
      and `GET /api/runs/{id}/steps`
- [ ] 2.2 `metrics.OutcomeOf` and the DTO fields

## 3. Frontend

- [ ] 3.1 Failing tests: the tone and label from an outcome, stage progress
      from step outcomes
- [ ] 3.2 Switch `statusModel`, `stageProgress`, the card and the page's
      polling to `outcome`; delete the string interpretation
- [ ] 3.3 Update the runs Playwright fixtures

## 4. Wrap up

- [ ] 4.1 Archive this change once merged; tick finding 1 on #376
