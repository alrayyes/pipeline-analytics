// Shared client for GET/PATCH /api/settings -- account_settings' server
// record. Every persisted store (theme, forge filter, the Pipelines
// filters) goes through this, rather than each rolling its own fetch, so
// the request/response shape only needs to be gotten right once.

import { refusal } from './apiError.js';

export interface SettingsValues {
	theme: 'light' | 'dark' | 'system';
	forgeFilter: 'all' | 'github' | 'forgejo';
	pipelinesHealthFilter: 'all' | 'healthy' | 'unhealthy';
	pipelinesRepoSelector: string;
	pipelinesSortOrder: 'name' | 'lastRun';
	telemetryWindow: '24h' | '7d' | '30d';
}

// GET /api/settings answers every setting plus the server's own `defaults`
// for each, so no store keeps a copy of a default that could drift from the
// server's (rules/frontend.md).
export interface ServerSettings extends SettingsValues {
	defaults: SettingsValues;
}

// The read for a caller that needs to know why it failed, such as a WebMCP
// tool: a refusal throws an ApiError. `signal` lets a caller bound how long
// it's willing to wait (#251).
export async function loadSettings(
	fetchFn: typeof fetch = fetch,
	signal?: AbortSignal,
): Promise<ServerSettings> {
	const res = await fetchFn('/api/settings', { signal });
	if (!res.ok) throw await refusal(res);

	return (await res.json()) as ServerSettings;
}

// Used by +layout.ts's load(), which supplies SvelteKit's own fetch (so
// this participates in the same request dedup/credentials handling as the
// rest of that load call) -- defaults to the global fetch for every other
// caller (a user-triggered patch, not a load). An aborted or refused request
// is the same null as any other failed fetch.
export async function fetchSettings(
	fetchFn: typeof fetch = fetch,
	signal?: AbortSignal,
): Promise<ServerSettings | null> {
	try {
		return await loadSettings(fetchFn, signal);
	} catch {
		return null;
	}
}

export type SettingsUpdate = Partial<
	Record<keyof SettingsValues, string | null>
>;

// The PATCH itself, answering every setting after the update. A refusal throws
// an ApiError carrying the server's message.
export async function saveSettings(
	update: SettingsUpdate,
	fetchFn: typeof fetch = fetch,
): Promise<ServerSettings> {
	const res = await fetchFn('/api/settings', {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(update),
	});
	if (!res.ok) throw await refusal(res);

	return (await res.json()) as ServerSettings;
}

// A key set to null clears it back to its documented default -- the same
// convention the PATCH endpoint itself uses (design.md's "null clears to
// default"). Best-effort: a store's own optimistic local update and
// localStorage cache write already happened by the time this is called, so
// a failed PATCH here just means the next full settings fetch (a reload,
// another device) is what eventually reconciles it.
export async function patchSettings(update: SettingsUpdate): Promise<void> {
	try {
		await saveSettings(update);
	} catch {
		// Best effort -- see doc comment above.
	}
}
