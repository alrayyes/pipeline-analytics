export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme';

function readStored(): Theme {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return stored === 'light' || stored === 'dark' ? stored : 'system';
	} catch {
		return 'system';
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

export function setTheme(next: Theme): void {
	theme = next;

	try {
		localStorage.setItem(STORAGE_KEY, next);
	} catch {
		// Best effort -- an unavailable localStorage (private browsing,
		// blocked storage) just means the choice doesn't persist.
	}

	apply(next);
}

export function cycleTheme(): void {
	setTheme(theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light');
}

// Called once from the root layout: syncs the in-app state with what
// app.html's inline script already applied before first paint, and keeps
// the effective theme in sync if the OS setting changes while "system" is
// selected.
export function initTheme(): () => void {
	theme = readStored();
	apply(theme);

	const media = window.matchMedia('(prefers-color-scheme: dark)');
	const onChange = () => {
		if (theme === 'system') apply(theme);
	};
	media.addEventListener('change', onChange);

	return () => media.removeEventListener('change', onChange);
}
