# Design

## Decisions

**Health is computed, so the filter can't be SQL.** A pipeline's health comes
from its recent runs and steps (`summarize`), not a stored column. Filtering
by it, or sorting by last run, therefore has to summarise every pipeline that
matches the repo and forge filters, then filter, sort and page the result. The
unfiltered, name-sorted path keeps its page-at-a-time query.

**Cost.** The whole-set path summarises every matching pipeline per request.
This app tracks a solo developer's repos, and the page previously did a full
unpaginated fetch for exactly this reason; the cost scales with tracked
pipelines, not runs. Worth revisiting if a deployment tracks thousands.

**Ordering.** `sort=name` is the existing order, by repository then name.
`sort=lastRun` orders by most recent run first across the whole set, a
pipeline with no runs last, ties by repository then name so the order is
stable.

**Invalid values are a 400**, as for the run list's `status`: a typo must not
read as "no unhealthy pipelines". An absent parameter means no filter and the
default order.

**The page keeps grouping.** Grouping pipelines under their repository is
presentation and stays in the browser, but it now follows the server's order
instead of re-sorting inside each group.

## Risks / Trade-offs

- **Per-request cost grows with tracked pipelines**, as above.
- **Changing the filter resets paging**, which the page already did for the
  repo and forge filters.
