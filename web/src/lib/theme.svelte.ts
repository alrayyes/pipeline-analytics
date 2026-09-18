import { patchSettings } from '$lib/settingsSync.js';

export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme';

// localStorage is a paint-only cache now (persist-account-settings/
// design.md), populated after every successful sync with the server --
// app.html's inline anti-flash script still reads this same key
// synchronously before hydration, which is the one read this cache exists
// for.
function readCached(): Theme {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return stored === 'light' || stored === 'dark' ? stored : 'system';
	} catch {
		return 'system';
	}
}

function writeCache(theme: Theme): void {
	try {
		localStorage.setItem(STORAGE_KEY, theme);
	} catch {
		// Best effort -- an unavailable localStorage (private browsing,
		// blocked storage) just means the choice doesn't persist locally;
		// the server round trip is what actually matters now.
	}
}

function prefersDark(): boolean {
	return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

function apply(theme: Theme): void {
	const dark = theme === 'dark' || (theme === 'system' && prefersDark());
	document.documentElement.classList.toggle('dark', dark);
}

// $state module-level runes are shared across every importer, which is
// exactly what a single global theme wants -- one source of truth read
// and written from both the layout and the toggle button.
let theme = $state<Theme>('system');

export function getTheme(): Theme {
	return theme;
}

// Applies and caches immediately (so the toggle feels instant and the next
// cold load paints correctly even if the PATCH below hasn't landed yet),
// then syncs to the server -- the actual source of truth from here on.
export function setTheme(next: Theme): void {
	theme = next;
	apply(next);
	writeCache(next);

	void patchSettings({ theme: next });
}

export function cycleTheme(): void {
	setTheme(theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light');
}

// Called once from the root layout, with the theme persist-account-
// settings' load() already fetched server-side (null if that fetch
// failed, or on a route -- /login -- that doesn't fetch it at all).
// Reconciles the in-app state with whatever app.html's inline script
// already applied before first paint: the server value if there is one
// (it might differ from the cache, e.g. a change made on another device),
// falling back to the same cached value app.html used otherwise. Also
// keeps the effective theme in sync if the OS setting changes while
// "system" is selected.
export function initTheme(serverTheme?: Theme): () => void {
	theme = serverTheme ?? readCached();
	apply(theme);
	writeCache(theme);

	const media = window.matchMedia('(prefers-color-scheme: dark)');
	const onChange = () => {
		if (theme === 'system') apply(theme);
	};
	media.addEventListener('change', onChange);

	return () => media.removeEventListener('change', onChange);
}
