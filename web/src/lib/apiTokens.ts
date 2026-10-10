// Client for POST /api/auth/tokens: an API token whose lifetime the user
// picks from presets. The presets are UX only -- the server clamps whatever
// it is sent to its own ceiling (#191) -- and none of them is "never
// expires", because every token expires.

const DAY_SECONDS = 86_400;

export type TtlPreset = '30d' | '90d' | '1y';

const TTL_SECONDS: Record<TtlPreset, number> = {
	'30d': 30 * DAY_SECONDS,
	'90d': 90 * DAY_SECONDS,
	'1y': 365 * DAY_SECONDS,
};

export const TTL_PRESETS: {
	value: TtlPreset;
	label: string;
	ttlSeconds: number;
}[] = [
	{ value: '30d', label: '30 days', ttlSeconds: TTL_SECONDS['30d'] },
	{ value: '90d', label: '90 days', ttlSeconds: TTL_SECONDS['90d'] },
	{ value: '1y', label: '1 year', ttlSeconds: TTL_SECONDS['1y'] },
];

// Not the longest: a short-lived CI credential is the common case.
export const DEFAULT_TTL_PRESET: TtlPreset = '90d';

export interface IssuedApiToken {
	id: string;
	token: string;
	createdAt: string;
	expiresAt: string;
}

// The expiry the server will apply to a token created at `now`, for the
// live preview before anything is created.
export function expiryFor(preset: TtlPreset, now: Date = new Date()): Date {
	return new Date(now.getTime() + TTL_SECONDS[preset] * 1000);
}

export async function createApiToken(
	preset: TtlPreset,
	fetchFn: typeof fetch = fetch,
): Promise<IssuedApiToken> {
	const res = await fetchFn('/api/auth/tokens', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ ttlSeconds: TTL_SECONDS[preset] }),
	});

	if (!res.ok) {
		// Stryker disable next-line ArrowFunction
		const body = await res.json().catch(() => null);

		throw new Error(body?.message ?? 'Could not create the token.');
	}

	return (await res.json()) as IssuedApiToken;
}
