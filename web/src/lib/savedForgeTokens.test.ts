import { describe, expect, test } from 'bun:test';
import {
	deleteForgeToken,
	findSavedToken,
	listForgeTokens,
	type SavedForgeToken,
	savedTokenLabel,
	saveForgeToken,
} from './savedForgeTokens.js';

interface Call {
	url: string;
	init?: RequestInit;
}

function respond(
	status: number,
	body?: unknown,
): { calls: Call[]; fetchFn: typeof fetch } {
	const calls: Call[] = [];
	const fetchFn = ((input: RequestInfo | URL, init?: RequestInit) => {
		calls.push({ url: String(input), init });

		return Promise.resolve(
			new Response(body === undefined ? null : JSON.stringify(body), {
				status,
			}),
		);
	}) as typeof fetch;

	return { calls, fetchFn };
}

const github: SavedForgeToken = {
	id: 't1',
	forge: 'github',
	tokenMasked: '****1234',
};
const forgejo: SavedForgeToken = {
	id: 't2',
	forge: 'forgejo',
	forgejoInstanceUrl: 'https://forgejo.example.com',
	tokenMasked: '****5678',
};

describe('listForgeTokens', () => {
	test('reads the masked tokens from GET /api/forge-tokens', async () => {
		const { calls, fetchFn } = respond(200, { tokens: [github, forgejo] });

		expect(await listForgeTokens(fetchFn)).toEqual([github, forgejo]);
		expect(calls).toEqual([{ url: '/api/forge-tokens', init: undefined }]);
	});

	test('throws on a failed response', async () => {
		const { fetchFn } = respond(500, {});

		await expect(listForgeTokens(fetchFn)).rejects.toThrow(
			'Could not load saved tokens.',
		);
	});
});

describe('saveForgeToken', () => {
	test('PUTs the forge, instance and token, and returns the masked one', async () => {
		const { calls, fetchFn } = respond(200, forgejo);

		const saved = await saveForgeToken(
			{
				forge: 'forgejo',
				forgejoInstanceUrl: 'https://forgejo.example.com',
				token: 'secret-5678',
			},
			fetchFn,
		);

		expect(saved).toEqual(forgejo);
		expect(calls).toHaveLength(1);
		expect(calls[0].url).toBe('/api/forge-tokens');
		expect(calls[0].init?.method).toBe('PUT');
		expect(new Headers(calls[0].init?.headers).get('Content-Type')).toBe(
			'application/json',
		);
		expect(JSON.parse(String(calls[0].init?.body))).toEqual({
			forge: 'forgejo',
			forgejoInstanceUrl: 'https://forgejo.example.com',
			token: 'secret-5678',
		});
	});

	test('sends no instance URL for GitHub', async () => {
		const { calls, fetchFn } = respond(200, github);

		await saveForgeToken(
			{ forge: 'github', forgejoInstanceUrl: 'https://ignored', token: 'x' },
			fetchFn,
		);

		expect(JSON.parse(String(calls[0].init?.body))).toEqual({
			forge: 'github',
			token: 'x',
		});
	});

	test("throws the server's own message when it refuses", async () => {
		const { fetchFn } = respond(400, {
			message: 'forgejoInstanceUrl is required',
		});

		await expect(
			saveForgeToken({ forge: 'forgejo', token: 'x' }, fetchFn),
		).rejects.toThrow('forgejoInstanceUrl is required');
	});

	test('falls back to a generic message when the refusal has none', async () => {
		const { fetchFn } = respond(500);

		await expect(
			saveForgeToken({ forge: 'github', token: 'x' }, fetchFn),
		).rejects.toThrow('Could not save the token.');
	});
});

describe('deleteForgeToken', () => {
	test('DELETEs the token by id', async () => {
		const { calls, fetchFn } = respond(204);

		await deleteForgeToken('t 1', fetchFn);

		expect(calls).toHaveLength(1);
		expect(calls[0].url).toBe('/api/forge-tokens/t%201');
		expect(calls[0].init?.method).toBe('DELETE');
	});

	test('throws on a failed response', async () => {
		const { fetchFn } = respond(404, {});

		await expect(deleteForgeToken('t1', fetchFn)).rejects.toThrow(
			'Could not delete the token.',
		);
	});
});

describe('findSavedToken', () => {
	const tokens = [github, forgejo];

	test('finds the GitHub token whatever instance URL is passed', () => {
		expect(findSavedToken(tokens, 'github', undefined)).toBe(github);
		expect(findSavedToken(tokens, 'github', 'https://x')).toBe(github);
	});

	test('finds a Forgejo token by instance, ignoring spacing and a trailing slash, as the server does', () => {
		expect(
			findSavedToken(tokens, 'forgejo', ' https://forgejo.example.com/ '),
		).toBe(forgejo);
	});

	test('ignores every trailing slash, as the server does', () => {
		expect(
			findSavedToken(tokens, 'forgejo', 'https://forgejo.example.com//'),
		).toBe(forgejo);
	});

	test('never matches a Forgejo token that has no instance to an empty or missing one', () => {
		const orphan: SavedForgeToken = {
			id: 't3',
			forge: 'forgejo',
			tokenMasked: '****0000',
		};

		expect(findSavedToken([orphan], 'forgejo', '')).toBe(undefined);
		expect(findSavedToken([orphan], 'forgejo', undefined)).toBe(undefined);
		expect(findSavedToken([orphan], 'forgejo', '  /')).toBe(undefined);
	});

	test('finds nothing for another instance, or a forge with no token', () => {
		expect(findSavedToken(tokens, 'forgejo', 'https://other.example')).toBe(
			undefined,
		);
		expect(findSavedToken(tokens, 'forgejo', '')).toBe(undefined);
		expect(findSavedToken([forgejo], 'github', undefined)).toBe(undefined);
	});
});

describe('savedTokenLabel', () => {
	test('names the forge, and the instance for Forgejo', () => {
		expect(savedTokenLabel(github)).toBe('GitHub');
		expect(savedTokenLabel(forgejo)).toBe(
			'Forgejo · https://forgejo.example.com',
		);
	});
});
