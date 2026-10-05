const LAST_SELECTION_KEY = 'lastForgeSelection';

interface ForgeSelection {
	forge: 'github' | 'forgejo';
	instanceUrl?: string;
}

// Remembers which forge (+ instance URL, for Forgejo) was last used to
// register a repo, so the register dialog starts there instead of always on
// GitHub -- otherwise the token saved for any other forge/instance wouldn't
// be the one in use (#262). Not a secret, so it stays browser-local; the
// tokens themselves are saved server-side (#462).
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
		return JSON.parse(
			localStorage.getItem(LAST_SELECTION_KEY) ?? 'null',
		) as ForgeSelection | null;
	} catch {
		return null;
	}
}

const LEGACY_TOKEN_PREFIX = 'rememberedToken:';

// Tokens used to be kept here in the clear (#72), where any script on the
// page can read them. They're saved server-side now, so the old copies are
// removed rather than left behind. Collects the keys first: removing while
// walking the index would skip entries.
export function purgeRememberedTokens(): void {
	try {
		const keys: string[] = [];

		for (let i = 0; i < localStorage.length; i++) {
			const key = localStorage.key(i);
			if (key?.startsWith(LEGACY_TOKEN_PREFIX)) keys.push(key);
		}

		for (const key of keys) localStorage.removeItem(key);
	} catch {
		// Best effort -- unavailable storage holds nothing to remove.
	}
}
