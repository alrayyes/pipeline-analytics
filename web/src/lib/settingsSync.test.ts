import { describe, expect, mock, test } from 'bun:test';
import { ApiError } from './apiError.js';
import {
	fetchSettings,
	loadSettings,
	patchSettings,
	saveSettings,
} from './settingsSync.js';

const settings = { theme: 'dark', defaults: { theme: 'system' } };

function fetchReturning(status: number, body?: unknown) {
	return mock((_url: string, _init?: RequestInit) =>
		Promise.resolve(
			new Response(body === undefined ? null : JSON.stringify(body), {
				status,
			}),
		),
	) as unknown as typeof fetch & ReturnType<typeof mock>;
}

describe('loadSettings', () => {
	test('reads every setting', async () => {
		const fetchFn = fetchReturning(200, settings);

		expect(await loadSettings(fetchFn)).toEqual(settings as never);
		expect(fetchFn.mock.calls[0]?.[0]).toBe('/api/settings');
	});

	test('throws an ApiError carrying the status', async () => {
		const error = await loadSettings(fetchReturning(401, {})).catch(
			(e: unknown) => e,
		);

		expect(error).toBeInstanceOf(ApiError);
		expect((error as ApiError).status).toBe(401);
	});
});

describe('saveSettings', () => {
	test('patches the update, nulls included, and returns the result', async () => {
		const fetchFn = fetchReturning(200, settings);

		const saved = await saveSettings(
			{ theme: 'dark', forgeFilter: null },
			fetchFn,
		);

		const [url, init] = fetchFn.mock.calls[0] ?? [];
		expect(url).toBe('/api/settings');
		expect(init?.method).toBe('PATCH');
		expect(init?.headers).toEqual({ 'Content-Type': 'application/json' });
		expect(init?.body).toBe(
			JSON.stringify({ theme: 'dark', forgeFilter: null }),
		);
		expect(saved).toEqual(settings as never);
	});

	test("keeps the server's message when it rejects the update", async () => {
		const error = await saveSettings(
			{ theme: 'blue' },
			fetchReturning(400, { code: 'bad_request', message: 'theme: invalid' }),
		).catch((e: unknown) => e as Error);

		expect(error).toBeInstanceOf(ApiError);
		expect(error.message).toBe(
			'Request failed with status 400: theme: invalid',
		);
	});
});

describe('the best-effort callers keep their old behaviour', () => {
	test('fetchSettings answers null on a refusal', async () => {
		expect(await fetchSettings(fetchReturning(500, {}))).toBeNull();
	});

	test('fetchSettings answers null when the request never arrives', async () => {
		const failing = (() => Promise.reject(new Error('down'))) as typeof fetch;

		expect(await fetchSettings(failing)).toBeNull();
	});

	test('patchSettings swallows a refusal', async () => {
		const original = globalThis.fetch;
		globalThis.fetch = fetchReturning(500, {});
		try {
			await expect(patchSettings({ theme: 'dark' })).resolves.toBeUndefined();
		} finally {
			globalThis.fetch = original;
		}
	});
});
