import { patchSettings } from '$lib/settingsSync.js';

export type ForgeFilter = 'all' | 'github' | 'forgejo';

const STORAGE_KEY = 'forgeFilter';

// localStorage is a paint-only cache now (persist-account-settings/
// design.md), populated after every successful sync with the server.
function readCached(): ForgeFilter {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return stored === 'github' || stored === 'forgejo' ? stored : 'all';
	} catch {
		return 'all';
	}
}

function writeCache(filter: ForgeFilter): void {
	try {
		localStorage.setItem(STORAGE_KEY, filter);
	} catch {
		// Best effort -- see theme.svelte.ts's identical comment.
	}
}

// $state module-level rune shared across every importer, which is exactly
// what a single filter applied consistently across pages wants -- one
// source of truth read and written from wherever the segmented control
// appears, same reasoning as theme.svelte.ts's own module-level theme.
let forgeFilter = $state<ForgeFilter>('all');

export function getForgeFilter(): ForgeFilter {
	return forgeFilter;
}

export function setForgeFilter(next: ForgeFilter): void {
	forgeFilter = next;
	writeCache(next);

	void patchSettings({ forgeFilter: next });
}

// Called once from the root layout, with the forge filter persist-account-
// settings' load() already fetched server-side (null if that fetch failed,
// or on a route that doesn't fetch it at all) -- falls back to the cached
// value otherwise, same reconciliation shape as theme.svelte.ts's
// initTheme.
export function initForgeFilter(serverFilter?: ForgeFilter): void {
	forgeFilter = serverFilter ?? readCached();
	writeCache(forgeFilter);
}
