import { patchSettings } from './settingsSync.js';

export type HealthFilter = 'all' | 'healthy' | 'unhealthy';
export type SortBy = 'name' | 'lastRun';

// Defaults to unhealthy-only (#101): this is a monitoring dashboard, so
// leading with what needs attention beats an everything-at-once list.
export const DEFAULT_HEALTH_FILTER: HealthFilter = 'unhealthy';
export const DEFAULT_REPO_SELECTOR = 'all';
export const DEFAULT_SORT_BY: SortBy = 'name';

const HEALTH_FILTER_KEY = 'pipelinesHealthFilter';
const REPO_SELECTOR_KEY = 'pipelinesRepoSelector';
const SORT_BY_KEY = 'pipelinesSortOrder';

// localStorage is a paint-only cache (persist-account-settings/design.md),
// populated after every successful sync with the server -- same shape as
// theme.svelte.ts/forgeFilter.svelte.ts, extended to the Pipelines page's
// three filter controls, which used to be plain, unpersisted component
// state.
function readCached(key: string): string | null {
	try {
		return localStorage.getItem(key);
	} catch {
		return null;
	}
}

function writeCache(key: string, value: string): void {
	try {
		localStorage.setItem(key, value);
	} catch {
		// Best effort -- see theme.svelte.ts's identical comment.
	}
}

let healthFilter = $state<HealthFilter>(DEFAULT_HEALTH_FILTER);
let repoSelector = $state(DEFAULT_REPO_SELECTOR);
let sortBy = $state<SortBy>(DEFAULT_SORT_BY);

export function getHealthFilter(): HealthFilter {
	return healthFilter;
}

export function setHealthFilter(next: HealthFilter): void {
	healthFilter = next;
	writeCache(HEALTH_FILTER_KEY, next);

	void patchSettings({ pipelinesHealthFilter: next });
}

export function getRepoSelector(): string {
	return repoSelector;
}

export function setRepoSelector(next: string): void {
	repoSelector = next;
	writeCache(REPO_SELECTOR_KEY, next);

	void patchSettings({ pipelinesRepoSelector: next });
}

export function getSortBy(): SortBy {
	return sortBy;
}

export function setSortBy(next: SortBy): void {
	sortBy = next;
	writeCache(SORT_BY_KEY, next);

	void patchSettings({ pipelinesSortOrder: next });
}

// Whether every filter already matches its default -- the Pipelines page's
// "Reset filters" control disables itself (not hides, matching this app's
// existing convention for a currently-inapplicable action) when this is
// already true.
export function isAtDefaults(): boolean {
	return (
		healthFilter === DEFAULT_HEALTH_FILTER &&
		repoSelector === DEFAULT_REPO_SELECTOR &&
		sortBy === DEFAULT_SORT_BY
	);
}

// Resets all three filters to their documented defaults in one request
// (account-settings/spec.md's "Reset Pipelines filters to defaults") --
// PATCHing all three keys to null rather than to their literal default
// values, so this stays correct even if a default ever changes server-side
// without this client also having to know the new value.
export function resetFilters(): void {
	healthFilter = DEFAULT_HEALTH_FILTER;
	repoSelector = DEFAULT_REPO_SELECTOR;
	sortBy = DEFAULT_SORT_BY;

	writeCache(HEALTH_FILTER_KEY, DEFAULT_HEALTH_FILTER);
	writeCache(REPO_SELECTOR_KEY, DEFAULT_REPO_SELECTOR);
	writeCache(SORT_BY_KEY, DEFAULT_SORT_BY);

	void patchSettings({
		pipelinesHealthFilter: null,
		pipelinesRepoSelector: null,
		pipelinesSortOrder: null,
	});
}

// Called once from the root layout, with the Pipelines filters persist-
// account-settings' load() already fetched server-side (undefined if that
// fetch failed, or on a route that doesn't fetch it at all) -- falls back
// to the cached values otherwise, same reconciliation shape as
// theme.svelte.ts's initTheme.
export function initPipelinesFilters(server?: {
	pipelinesHealthFilter: HealthFilter;
	pipelinesRepoSelector: string;
	pipelinesSortOrder: SortBy;
}): void {
	healthFilter =
		server?.pipelinesHealthFilter ??
		(readCached(HEALTH_FILTER_KEY) as HealthFilter | null) ??
		DEFAULT_HEALTH_FILTER;
	repoSelector =
		server?.pipelinesRepoSelector ??
		readCached(REPO_SELECTOR_KEY) ??
		DEFAULT_REPO_SELECTOR;
	sortBy =
		server?.pipelinesSortOrder ??
		(readCached(SORT_BY_KEY) as SortBy | null) ??
		DEFAULT_SORT_BY;

	writeCache(HEALTH_FILTER_KEY, healthFilter);
	writeCache(REPO_SELECTOR_KEY, repoSelector);
	writeCache(SORT_BY_KEY, sortBy);
}
