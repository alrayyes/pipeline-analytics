const STORAGE_PREFIX = 'rememberedToken:';

function storageKey(forge: string, instanceUrl?: string): string {
	return `${STORAGE_PREFIX}${forge}:${instanceUrl?.trim() ?? ''}`;
}

// Remembers a token per (forge, instance URL) in the browser only -- there's
// no server-side credential entity, so this doesn't survive a cleared
// browser or work across devices, an accepted tradeoff for not needing a
// schema change (issue #72's design decision).
export function rememberToken(
	forge: string,
	instanceUrl: string | undefined,
	token: string,
): void {
	try {
		localStorage.setItem(storageKey(forge, instanceUrl), token);
	} catch {
		// Best effort -- an unavailable localStorage (private browsing,
		// blocked storage) just means the token isn't offered next time.
	}
}

export function getRememberedToken(
	forge: string,
	instanceUrl?: string,
): string | null {
	try {
		return localStorage.getItem(storageKey(forge, instanceUrl));
	} catch {
		return null;
	}
}

// Mirrors the backend's MaskToken (internal/ingestion/ingestion.go) so a
// remembered token reads the same way a registered repo's does.
export function maskToken(token: string): string {
	const visible = 4;
	if (token.length <= visible) return '****';

	return `****${token.slice(-visible)}`;
}

const LAST_SELECTION_KEY = `${STORAGE_PREFIX}lastSelection`;

interface ForgeSelection {
	forge: 'github' | 'forgejo';
	instanceUrl?: string;
}

// Remembers which forge (+ instance URL, for Forgejo) was last used to
// register a repo, so the register dialog can restore it as its starting
// selection instead of always resetting to GitHub -- otherwise a remembered
// token for any other forge/instance never gets looked up under the right
// key (#262). Same browser-local, no-server-entity tradeoff as the token
// itself.
export function rememberForgeSelection(
	forge: 'github' | 'forgejo',
	instanceUrl: string | undefined,
): void {
	try {
		const selection: ForgeSelection =
			forge === 'forgejo' && instanceUrl ? { forge, instanceUrl } : { forge };
		localStorage.setItem(LAST_SELECTION_KEY, JSON.stringify(selection));
	} catch {
		// Best effort -- the dialog just falls back to the GitHub default.
	}
}

export function getRememberedForgeSelection(): ForgeSelection | null {
	try {
		const raw = localStorage.getItem(LAST_SELECTION_KEY);
		if (!raw) return null;

		return JSON.parse(raw) as ForgeSelection;
	} catch {
		return null;
	}
}
