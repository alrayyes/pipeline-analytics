import {
	patchSettings,
	type ServerSettings,
	type SettingsValues,
} from './settingsSync.js';

export type HealthFilter = 'all' | 'healthy' | 'unhealthy';
export type SortBy = 'name' | 'lastRun';

// What the filters show while neither the server nor the cache has said
// anything: everything, unfiltered. It is not a default -- the server owns
// those and sends them as `defaults` (rules/frontend.md) -- so it never
// feeds isAtDefaults() or resetFilters().
const UNFILTERED_HEALTH: HealthFilter = 'all';
const UNFILTERED_REPO = 'all';
const UNSORTED: SortBy = 'name';

const HEALTH_FILTER_KEY = 'pipelinesHealthFilter';
const REPO_SELECTOR_KEY = 'pipelinesRepoSelector';
const SORT_BY_KEY = 'pipelinesSortOrder';

// localStorage is a paint-only cache (persist-account-settings/design.md),
// populated after every successful sync with the server -- same shape as
// theme.svelte.ts/forgeFilter.svelte.ts, extended to the Pipelines page's
// three filter controls, which used to be plain, unpersisted component
// state.
// An empty catch returns undefined, and every caller treats that like null
// (`??`), so that mutant is equivalent and ignored for this function.
// Stryker disable BlockStatement
function readCached(key: string): string | null {
	try {
		return localStorage.getItem(key);
	} catch {
		return null;
	}
}
// Stryker restore BlockStatement

function writeCache(key: string, value: string): void {
	try {
		localStorage.setItem(key, value);
	} catch {
		// Best effort -- see theme.svelte.ts's identical comment.
	}
}

let healthFilter = $state<HealthFilter>(UNFILTERED_HEALTH);
let repoSelector = $state(UNFILTERED_REPO);
let sortBy = $state<SortBy>(UNSORTED);

// The server's defaults, from the last settings response; undefined until
// one arrives (a failed fetch, or a route that doesn't make it).
let defaults = $state<SettingsValues | undefined>();

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

// Whether every filter already matches the server's default -- the
// Pipelines page's "Reset filters" control disables itself (not hides,
// matching this app's existing convention for a currently-inapplicable
// action) when this is true. With no defaults to compare against there is
// nothing to reset to, so it reports true and the control stays disabled.
export function isAtDefaults(): boolean {
	if (!defaults) return true;

	return (
		healthFilter === defaults.pipelinesHealthFilter &&
		repoSelector === defaults.pipelinesRepoSelector &&
		sortBy === defaults.pipelinesSortOrder
	);
}

// Resets all three filters to the server's defaults in one request
// (account-settings/spec.md's "Reset Pipelines filters to defaults"). The
// PATCH sends null rather than the default values, so the server stays the
// one place that decides what they are; the local update only mirrors its
// answer, and is skipped when the defaults are unknown.
export function resetFilters(): void {
	if (defaults) {
		healthFilter = defaults.pipelinesHealthFilter;
		repoSelector = defaults.pipelinesRepoSelector;
		sortBy = defaults.pipelinesSortOrder;

		writeCache(HEALTH_FILTER_KEY, healthFilter);
		writeCache(REPO_SELECTOR_KEY, repoSelector);
		writeCache(SORT_BY_KEY, sortBy);
	}

	void patchSettings({
		pipelinesHealthFilter: null,
		pipelinesRepoSelector: null,
		pipelinesSortOrder: null,
	});
}

// Called once from the root layout, with the settings persist-account-
// settings' load() already fetched server-side (undefined if that fetch
// failed, or on a route that doesn't fetch it at all) -- falls back to the
// cached values otherwise, same reconciliation shape as theme.svelte.ts's
// initTheme.
export function initPipelinesFilters(server?: ServerSettings): void {
	defaults = server?.defaults;

	healthFilter =
		server?.pipelinesHealthFilter ??
		(readCached(HEALTH_FILTER_KEY) as HealthFilter | null) ??
		UNFILTERED_HEALTH;
	repoSelector =
		server?.pipelinesRepoSelector ??
		readCached(REPO_SELECTOR_KEY) ??
		UNFILTERED_REPO;
	sortBy =
		server?.pipelinesSortOrder ??
		(readCached(SORT_BY_KEY) as SortBy | null) ??
		UNSORTED;

	writeCache(HEALTH_FILTER_KEY, healthFilter);
	writeCache(REPO_SELECTOR_KEY, repoSelector);
	writeCache(SORT_BY_KEY, sortBy);
}
