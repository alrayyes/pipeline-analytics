// Client for /api/forge-tokens: the forge tokens saved for registering
// repositories. The server returns only the masked form, never the token
// (#462), so nothing here can show or reuse one -- registration and
// discovery are sent with no `token` and the server uses the saved one.

export interface SavedForgeToken {
	id: string;
	forge: 'github' | 'forgejo';
	forgejoInstanceUrl?: string;
	tokenMasked: string;
}

export interface SaveForgeTokenInput {
	forge: 'github' | 'forgejo';
	forgejoInstanceUrl?: string;
	token: string;
}

export async function listForgeTokens(
	fetchFn: typeof fetch = fetch,
): Promise<SavedForgeToken[]> {
	const res = await fetchFn('/api/forge-tokens');
	if (!res.ok) throw new Error('Could not load saved tokens.');

	return ((await res.json()) as { tokens: SavedForgeToken[] }).tokens;
}

// Replaces the token saved for that forge and instance. Throws the server's
// own message when it refuses, so the form can show why.
export async function saveForgeToken(
	input: SaveForgeTokenInput,
	fetchFn: typeof fetch = fetch,
): Promise<SavedForgeToken> {
	const res = await fetchFn('/api/forge-tokens', {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			forge: input.forge,
			...(input.forge === 'forgejo'
				? { forgejoInstanceUrl: input.forgejoInstanceUrl }
				: {}),
			token: input.token,
		}),
	});

	if (!res.ok) {
		// An unreadable body and a body with no message read the same below.
		// Stryker disable next-line ArrowFunction
		const body = await res.json().catch(() => null);

		throw new Error(body?.message ?? 'Could not save the token.');
	}

	return (await res.json()) as SavedForgeToken;
}

export async function deleteForgeToken(
	id: string,
	fetchFn: typeof fetch = fetch,
): Promise<void> {
	const res = await fetchFn(`/api/forge-tokens/${encodeURIComponent(id)}`, {
		method: 'DELETE',
	});
	if (!res.ok) throw new Error('Could not delete the token.');
}

// The server matches an instance URL after trimming spaces and a trailing
// slash, so this does the same -- a looser match here would say "saved token
// in use" for a registration the server then refuses.
function instanceKey(url: string | undefined): string {
	return (url ?? '').trim().replace(/\/+$/, '');
}

export function findSavedToken(
	tokens: SavedForgeToken[],
	forge: 'github' | 'forgejo',
	instanceUrl: string | undefined,
): SavedForgeToken | undefined {
	return tokens.find(
		(t) =>
			t.forge === forge &&
			(forge === 'github' ||
				(instanceKey(instanceUrl) !== '' &&
					instanceKey(t.forgejoInstanceUrl) === instanceKey(instanceUrl))),
	);
}

export function savedTokenLabel(token: SavedForgeToken): string {
	return token.forge === 'forgejo'
		? `Forgejo · ${token.forgejoInstanceUrl}`
		: 'GitHub';
}
