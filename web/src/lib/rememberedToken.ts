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
