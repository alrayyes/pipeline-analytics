export type ForgeFilter = 'all' | 'github' | 'forgejo';

const STORAGE_KEY = 'forgeFilter';

function readStored(): ForgeFilter {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return stored === 'github' || stored === 'forgejo' ? stored : 'all';
	} catch {
		return 'all';
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

	try {
		localStorage.setItem(STORAGE_KEY, next);
	} catch {
		// Best effort -- an unavailable localStorage (private browsing,
		// blocked storage) just means the choice doesn't persist.
	}
}

// Called once from the root layout: hydrates the in-app state from
// whatever was last persisted, before any page reads it.
export function initForgeFilter(): void {
	forgeFilter = readStored();
}
