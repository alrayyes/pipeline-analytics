import { describe, expect, mock, test } from 'bun:test';
import { ApiError } from './dashboardApi.js';
import {
	discoverRepos,
	fetchRepoIdentifiers,
	fetchRepoPage,
	RepoApiError,
	registerRepo,
	untrackRepo,
} from './repoApi.js';

function fetchReturning(status: number, body?: unknown) {
	return mock((_url: string, _init?: RequestInit) =>
		Promise.resolve(
			new Response(body === undefined ? null : JSON.stringify(body), {
				status,
			}),
		),
	) as unknown as typeof fetch & ReturnType<typeof mock>;
}

describe('fetchRepoPage', () => {
	test('asks for one page, filtered by forge when given', async () => {
		const fetchFn = fetchReturning(200, { repos: [], hasMore: true });

		const page = await fetchRepoPage(
			{ limit: 20, offset: 40, forge: 'forgejo' },
			fetchFn,
		);

		expect(fetchFn.mock.calls[0]?.[0]).toBe(
			'/api/repos?limit=20&offset=40&forge=forgejo',
		);
		expect(page).toEqual({ repos: [], hasMore: true });
	});

	test('leaves the forge out when there is no filter', async () => {
		const fetchFn = fetchReturning(200, { repos: [], hasMore: false });

		await fetchRepoPage({ limit: 5, offset: 0 }, fetchFn);

		expect(fetchFn.mock.calls[0]?.[0]).toBe('/api/repos?limit=5&offset=0');
	});

	test('throws an ApiError carrying the status', async () => {
		const error = await fetchRepoPage(
			{ limit: 1, offset: 0 },
			fetchReturning(401, {}),
		).catch((e: unknown) => e);

		expect(error).toBeInstanceOf(ApiError);
		expect((error as ApiError).status).toBe(401);
	});
});

describe('fetchRepoIdentifiers', () => {
	test('sends the forge and, for Forgejo, the instance', async () => {
		const fetchFn = fetchReturning(200, { identifiers: ['a/b'] });

		const found = await fetchRepoIdentifiers(
			{ forge: 'forgejo', forgejoInstanceUrl: 'https://git.example' },
			fetchFn,
		);

		expect(fetchFn.mock.calls[0]?.[0]).toBe(
			'/api/repos/identifiers?forge=forgejo&forgejoInstanceUrl=https%3A%2F%2Fgit.example',
		);
		expect(found).toEqual(['a/b']);
	});

	test('sends only the forge when there is no instance', async () => {
		const fetchFn = fetchReturning(200, { identifiers: [] });

		await fetchRepoIdentifiers({ forge: 'github' }, fetchFn);

		expect(fetchFn.mock.calls[0]?.[0]).toBe(
			'/api/repos/identifiers?forge=github',
		);
	});
});

describe('discoverRepos', () => {
	test('posts the forge, the instance and the token', async () => {
		const fetchFn = fetchReturning(200, ['a/b']);

		const found = await discoverRepos(
			{ forge: 'forgejo', forgejoInstanceUrl: 'https://g.x', token: 't' },
			fetchFn,
		);

		const [url, init] = fetchFn.mock.calls[0] ?? [];
		expect(url).toBe('/api/repos/discover');
		expect(init?.method).toBe('POST');
		expect(init?.headers).toEqual({ 'Content-Type': 'application/json' });
		expect(JSON.parse(init?.body as string)).toEqual({
			forge: 'forgejo',
			forgejoInstanceUrl: 'https://g.x',
			token: 't',
		});
		expect(found).toEqual(['a/b']);
	});

	test('sends no token key when there is none, so the saved one is used', async () => {
		const fetchFn = fetchReturning(200, []);

		await discoverRepos({ forge: 'github' }, fetchFn);

		const init = fetchFn.mock.calls[0]?.[1];
		expect(JSON.parse(init?.body as string)).toEqual({ forge: 'github' });
	});

	test("keeps the server's code and message on a refusal", async () => {
		const error = await discoverRepos(
			{ forge: 'github' },
			fetchReturning(400, { code: 'no_saved_token', message: 'No token.' }),
		).catch((e: unknown) => e);

		expect(error).toBeInstanceOf(RepoApiError);
		expect(error).toMatchObject({
			status: 400,
			code: 'no_saved_token',
			serverMessage: 'No token.',
		});
	});

	test('a refusal with no JSON body still carries the status', async () => {
		const error = await discoverRepos(
			{ forge: 'github' },
			fetchReturning(502),
		).catch((e: unknown) => e);

		expect(error).toMatchObject({ status: 502 });
		expect((error as RepoApiError).code).toBeUndefined();
		expect((error as RepoApiError).serverMessage).toBeUndefined();
	});
});

describe('registerRepo', () => {
	test('posts the registration and returns the repo', async () => {
		const repo = { id: 'r1', forge: 'github', identifier: 'a/b' };
		const fetchFn = fetchReturning(201, repo);

		const registered = await registerRepo(
			{ forge: 'github', identifier: 'a/b', token: 't' },
			fetchFn,
		);

		const [url, init] = fetchFn.mock.calls[0] ?? [];
		expect(url).toBe('/api/repos');
		expect(init?.method).toBe('POST');
		expect(JSON.parse(init?.body as string)).toEqual({
			forge: 'github',
			identifier: 'a/b',
			token: 't',
		});
		expect(registered).toEqual(repo);
	});

	test('an already tracked repo is a 409 with its code', async () => {
		const error = await registerRepo(
			{ forge: 'github', identifier: 'a/b' },
			fetchReturning(409, { code: 'already_tracked', message: 'dup' }),
		).catch((e: unknown) => e);

		expect(error).toMatchObject({ status: 409, code: 'already_tracked' });
	});
});

describe('untrackRepo', () => {
	test('sends a DELETE for the repo', async () => {
		const fetchFn = fetchReturning(204);

		await untrackRepo('r 1', fetchFn);

		const [url, init] = fetchFn.mock.calls[0] ?? [];
		expect(url).toBe('/api/repos/r%201');
		expect(init?.method).toBe('DELETE');
	});

	test('throws an ApiError when the repo is gone', async () => {
		const error = await untrackRepo('r', fetchReturning(404, {})).catch(
			(e: unknown) => e,
		);

		expect(error).toBeInstanceOf(ApiError);
		expect((error as ApiError).status).toBe(404);
	});
});
