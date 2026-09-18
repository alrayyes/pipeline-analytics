# Tasks

## 1. Shared pipeline-list component

- [ ] 1.1 Extract today's homepage body (health filter, sort, pipeline
      cards) into a reusable component taking an optional `repoId`,
      and verify the existing homepage still renders identically
      through it before moving on

## 2. Repo overview (new `/`)

- [ ] 2.1 Add the repo health-rollup logic (worst-of aggregation:
      unhealthy if any pipeline is, else healthy, else "no runs yet")
      with unit tests covering all three states
- [ ] 2.2 Build the repo overview page: cards with aggregate status,
      unhealthy pipelines named inline, all/healthy/unhealthy filter
      defaulting to unhealthy, and verify against
      `specs/dashboard-ui/spec.md`'s two "Repo/pipeline overview"
      scenarios
- [ ] 2.3 Replace `web/src/routes/+page.svelte`'s flat pipeline list
      with the repo overview

## 3. Repo-scoped pipeline list (new `/repos/[id]`)

- [ ] 3.1 Add `web/src/routes/repos/[id]/+page.svelte` using the
      shared component from 1.1, scoped to `page.params.id`, and
      verify it matches the "Repo-scoped pipeline list" requirement's
      scenario
- [ ] 3.2 Verify navigating from a repo overview card lands on the
      correct repo's pipeline list

## 4. Repo management moves into Settings

- [ ] 4.1 Extract `web/src/routes/repos/+page.svelte`'s discover/
      add/untrack body into `RepoManagement.svelte`
- [ ] 4.2 Mount it as a "Repositories" section on `/settings`
      (requires `persist-account-settings`'s settings page to exist)
      and verify discover/add/untrack still work from there
- [ ] 4.3 Remove `web/src/routes/repos/+page.svelte`

## 5. Navigation

- [ ] 5.1 Update `Nav.svelte`: remove "Pipelines" and "Repositories"
      links, add "Settings", and verify the logo link still reaches
      the (now repo-overview) homepage

## 6. Verification

- [ ] 6.1 Update the Playwright suite (including the axe-core scan)
      for the new page structure and verify it passes
- [ ] 6.2 Update `README.md` if it documents the old page structure or
      nav, and verify it reads true again

## 7. Ship it

- [ ] 7.1 Open a pull request with `Closes #195` and verify CI passes
