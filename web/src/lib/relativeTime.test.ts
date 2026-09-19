import { describe, expect, test } from 'bun:test';
import { formatRelativeTime } from './relativeTime.js';

const NOW = new Date('2026-09-19T12:00:00Z');

describe('formatRelativeTime', () => {
	test('under a minute still renders in minutes, never seconds', () => {
		const date = new Date(NOW.getTime() - 30 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('this minute');
	});

	test('picks minutes as the coarsest unit under an hour', () => {
		const date = new Date(NOW.getTime() - 5 * 60 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('5 minutes ago');
	});

	test('picks hours once at least one whole hour has passed', () => {
		const date = new Date(NOW.getTime() - 3 * 60 * 60 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('3 hours ago');
	});

	test('picks days once at least one whole day has passed', () => {
		const date = new Date(NOW.getTime() - 2 * 86400 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('2 days ago');
	});

	test('a future date renders in the future tense', () => {
		const date = new Date(NOW.getTime() + 5 * 60 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('in 5 minutes');
	});
});
