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

	test('an exact unit boundary already counts as that unit', () => {
		const date = new Date(NOW.getTime() - 60 * 60 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('1 hour ago');
	});

	test('picks days once at least one whole day has passed', () => {
		const date = new Date(NOW.getTime() - 2 * 86400 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('2 days ago');
	});

	test('picks weeks once at least one whole week has passed', () => {
		const date = new Date(NOW.getTime() - 2 * 604800 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('2 weeks ago');
	});

	test('picks months once at least one whole month has passed', () => {
		const date = new Date(NOW.getTime() - 2 * 2592000 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('2 months ago');
	});

	test('picks years once at least one whole year has passed', () => {
		const date = new Date(NOW.getTime() - 2 * 31536000 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('2 years ago');
	});

	test('a future date renders in the future tense', () => {
		const date = new Date(NOW.getTime() + 5 * 60 * 1000);
		expect(formatRelativeTime(date, NOW)).toBe('in 5 minutes');
	});
});
