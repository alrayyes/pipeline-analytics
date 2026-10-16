// The Repositories page's requests: listing, registering, discovering and
// untracking repos. They live here, not inline in the page, so the page and
// the WebMCP tools send the same request (rules/mcp.md). A refusal throws an
// ApiError that also keeps the server's `code` and `message`, which the page
// words for the person; a request that never arrives rejects as fetch does.

import { ApiError } from './dashboardApi.js';

export type Forge = 'github' | 'forgejo';

export interface TrackedRepo {
	id: string;
	forge: Forge;
	identifier: string;
	forgejoInstanceUrl?: string;
	tokenMasked: string;
	ingestionStatus: 'pending' | 'active' | 'degraded';
	ingestionStatusReason?: string;
}

export interface RepoPage {
	repos: TrackedRepo[];
	hasMore: boolean;
}

export class RepoApiError extends ApiError {
	code?: string;
	serverMessage?: string;

	constructor(status: number, code?: string, serverMessage?: string) {
		super(status);
		this.code = code;
		this.serverMessage = serverMessage;
	}
}

async function refusal(res: Response): Promise<RepoApiError> {
	const body = (await res.json().catch(() => null)) as {
		code?: string;
		message?: string;
	} | null;

	return new RepoApiError(res.status, body?.code, body?.message);
}

const JSON_HEADERS = { 'Content-Type': 'application/json' };

export async function fetchRepoPage(
	params: { limit: number; offset: number; forge?: Forge },
	fetchFn: typeof fetch = fetch,
): Promise<RepoPage> {
	const query = new URLSearchParams({
		limit: String(params.limit),
		offset: String(params.offset),
	});
	if (params.forge) query.set('forge', params.forge);

	const res = await fetchFn(`/api/repos?${query}`);
	if (!res.ok) throw await refusal(res);

	return (await res.json()) as RepoPage;
}

// Every tracked identifier, unpaginated: Discover's "already tracked" check
// cannot rely on whichever page of the list is showing (#197).
export async function fetchRepoIdentifiers(
	params: { forge: Forge; forgejoInstanceUrl?: string },
	fetchFn: typeof fetch = fetch,
): Promise<string[]> {
	const query = new URLSearchParams({ forge: params.forge });
	if (params.forgejoInstanceUrl) {
		query.set('forgejoInstanceUrl', params.forgejoInstanceUrl);
	}

	const res = await fetchFn(`/api/repos/identifiers?${query}`);
	if (!res.ok) throw await refusal(res);

	return ((await res.json()) as { identifiers: string[] }).identifiers;
}

// A request without a `token` uses the one saved for the forge and instance.
export async function discoverRepos(
	request: { forge: Forge; forgejoInstanceUrl?: string; token?: string },
	fetchFn: typeof fetch = fetch,
): Promise<string[]> {
	const res = await fetchFn('/api/repos/discover', {
		method: 'POST',
		headers: JSON_HEADERS,
		body: JSON.stringify(request),
	});
	if (!res.ok) throw await refusal(res);

	return (await res.json()) as string[];
}

export async function registerRepo(
	request: {
		forge: Forge;
		identifier: string;
		forgejoInstanceUrl?: string;
		token?: string;
	},
	fetchFn: typeof fetch = fetch,
): Promise<TrackedRepo> {
	const res = await fetchFn('/api/repos', {
		method: 'POST',
		headers: JSON_HEADERS,
		body: JSON.stringify(request),
	});
	if (!res.ok) throw await refusal(res);

	return (await res.json()) as TrackedRepo;
}

export async function untrackRepo(
	repoId: string,
	fetchFn: typeof fetch = fetch,
): Promise<void> {
	const res = await fetchFn(`/api/repos/${encodeURIComponent(repoId)}`, {
		method: 'DELETE',
	});
	if (!res.ok) throw await refusal(res);
}
