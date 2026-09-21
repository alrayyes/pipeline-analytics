import { describe, expect, test } from 'bun:test';
import { isRedirect } from '@sveltejs/kit';
import { load } from './+layout.js';

type LoadFn = typeof load;
type LoadEvent = Parameters<LoadFn>[0];
// SvelteKit's `fetch` in a load event is structurally `typeof fetch`, which
// (depending on lib.dom's version) can require static properties like
// `preconnect` no test double bothers implementing -- these mocks only need
// to satisfy the plain `(input, init?) => Promise<Response>` call shape.
type FetchMock = (input: RequestInfo | URL) => Promise<Response>;

function callLoad(pathname: string, fetchFn: FetchMock) {
	return load({
		url: new URL(`http://localhost${pathname}`),
		fetch: fetchFn as unknown as LoadEvent['fetch'],
	} as LoadEvent);
}

describe('root layout load', () => {
	test('skips both requests on /login', async () => {
		const fetchFn: FetchMock = () => {
			throw new Error('should not be called on /login');
		};

		await expect(callLoad('/login', fetchFn)).resolves.toEqual({
			hasRepos: false,
			settings: null,
		});
	});

	test('reports hasRepos from a successful repos check', async () => {
		const fetchFn: FetchMock = (input) => {
			const url = String(input);
			if (url.includes('/api/repos')) {
				return Promise.resolve(
					new Response(JSON.stringify({ repos: [{ id: '1' }] }), {
						status: 200,
					}),
				);
			}
			return Promise.resolve(new Response(null, { status: 404 }));
		};

		await expect(callLoad('/', fetchFn)).resolves.toEqual({
			hasRepos: true,
			settings: null,
		});
	});

	test('redirects to /login on a 401', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 401 }));

		try {
			await callLoad('/', fetchFn);
			throw new Error('expected load() to redirect');
		} catch (e) {
			if (!isRedirect(e)) throw e;
			expect(e.status).toBe(302);
			expect(e.location).toBe('/login');
		}
	});

	// #251: a stalled/failed repos check used to have nothing to fall back
	// to -- it's now raced against a timeout and the fetch call itself is
	// wrapped in .catch(() => null), so a rejected fetch (a real network
	// failure, or the AbortSignal.timeout firing) degrades to hasRepos:
	// false instead of throwing and wedging the render.
	test('falls back to hasRepos: false when the repos check fails outright', async () => {
		const fetchFn: FetchMock = (input) => {
			const url = String(input);
			if (url.includes('/api/repos')) {
				return Promise.reject(new Error('network error'));
			}
			return Promise.resolve(new Response(null, { status: 404 }));
		};

		await expect(callLoad('/', fetchFn)).resolves.toEqual({
			hasRepos: false,
			settings: null,
		});
	});
});
