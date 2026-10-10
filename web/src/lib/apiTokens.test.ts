import { describe, expect, test } from 'bun:test';
import {
	createApiToken,
	DEFAULT_TTL_PRESET,
	expiryFor,
	TTL_PRESETS,
} from './apiTokens.js';

const DAY = 86_400;

function respond(
	status: number,
	body?: unknown,
): { calls: { url: string; init?: RequestInit }[]; fetchFn: typeof fetch } {
	const calls: { url: string; init?: RequestInit }[] = [];
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

describe('TTL_PRESETS', () => {
	test('offers 30 days, 90 days and 1 year, and nothing that never expires', () => {
		expect(TTL_PRESETS).toEqual([
			{ value: '30d', label: '30 days', ttlSeconds: 30 * DAY },
			{ value: '90d', label: '90 days', ttlSeconds: 90 * DAY },
			{ value: '1y', label: '1 year', ttlSeconds: 365 * DAY },
		]);
	});

	test('selects 90 days by default, not the longest', () => {
		expect(DEFAULT_TTL_PRESET).toBe('90d');
	});
});

describe('expiryFor', () => {
	test('is the preset lifetime after the given moment', () => {
		const now = new Date('2026-10-10T12:00:00Z');

		expect(expiryFor('30d', now).toISOString()).toBe(
			'2026-11-09T12:00:00.000Z',
		);
		expect(expiryFor('1y', now).toISOString()).toBe('2027-10-10T12:00:00.000Z');
	});
});

describe('createApiToken', () => {
	const issued = {
		id: 't1',
		token: 'pat_secret',
		createdAt: '2026-10-10T12:00:00Z',
		expiresAt: '2027-01-08T12:00:00Z',
	};

	test('posts the preset lifetime as ttlSeconds and returns the token', async () => {
		const { calls, fetchFn } = respond(201, issued);

		const token = await createApiToken('30d', fetchFn);

		expect(token).toEqual(issued);
		expect(calls).toHaveLength(1);
		const init = calls[0]?.init;
		expect(calls[0]?.url).toBe('/api/auth/tokens');
		expect(init?.method).toBe('POST');
		expect(new Headers(init?.headers).get('Content-Type')).toBe(
			'application/json',
		);
		expect(JSON.parse(String(init?.body))).toEqual({
			ttlSeconds: 30 * DAY,
		});
	});

	test('throws the server message when it refuses', async () => {
		const { fetchFn } = respond(400, {
			message: 'ttlSeconds must be positive',
		});

		await expect(createApiToken('90d', fetchFn)).rejects.toThrow(
			'ttlSeconds must be positive',
		);
	});

	test('falls back to a generic message when the body has none', async () => {
		const { fetchFn } = respond(500);

		await expect(createApiToken('90d', fetchFn)).rejects.toThrow(
			'Could not create the token.',
		);
	});
});
