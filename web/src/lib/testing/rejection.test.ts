import { describe, expect, test } from 'bun:test';
import { rejection } from './rejection.js';

describe('rejection', () => {
	test('returns what the promise rejected with', async () => {
		const boom = new Error('boom');

		expect(await rejection(Promise.reject(boom))).toBe(boom);
	});

	test('fails when the promise resolves', async () => {
		await expect(rejection(Promise.resolve(1))).rejects.toThrow(
			'expected the promise to reject',
		);
	});
});
