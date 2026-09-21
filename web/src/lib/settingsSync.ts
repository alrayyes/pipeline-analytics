// Shared client for GET/PATCH /api/settings -- account_settings' server
// record. Every persisted store (theme, forge filter, the Pipelines
// filters) goes through this, rather than each rolling its own fetch, so
// the request/response shape only needs to be gotten right once.

export interface ServerSettings {
	theme: 'light' | 'dark' | 'system';
	forgeFilter: 'all' | 'github' | 'forgejo';
	pipelinesHealthFilter: 'all' | 'healthy' | 'unhealthy';
	pipelinesRepoSelector: string;
	pipelinesSortOrder: 'name' | 'lastRun';
}

// Used by +layout.ts's load(), which supplies SvelteKit's own fetch (so
// this participates in the same request dedup/credentials handling as the
// rest of that load call) -- defaults to the global fetch for every other
// caller (a user-triggered patch, not a load). `signal` lets a caller bound
// how long it's willing to wait (#251) -- an aborted request rejects, which
// the catch below already treats the same as any other failed fetch.
export async function fetchSettings(
	fetchFn: typeof fetch = fetch,
	signal?: AbortSignal,
): Promise<ServerSettings | null> {
	try {
		const res = await fetchFn('/api/settings', { signal });
		if (!res.ok) return null;

		return (await res.json()) as ServerSettings;
	} catch {
		return null;
	}
}

// A key set to null clears it back to its documented default -- the same
// convention the PATCH endpoint itself uses (design.md's "null clears to
// default"). Best-effort: a store's own optimistic local update and
// localStorage cache write already happened by the time this is called, so
// a failed PATCH here just means the next full settings fetch (a reload,
// another device) is what eventually reconciles it.
export async function patchSettings(
	update: Partial<Record<keyof ServerSettings, string | null>>,
): Promise<void> {
	try {
		await fetch('/api/settings', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(update),
		});
	} catch {
		// Best effort -- see doc comment above.
	}
}
