import { describe, expect, test } from 'bun:test';
import { findRedirected } from './check-lighthouse-reports.ts';

describe('findRedirected', () => {
	test('passes when every report ended on the page it asked for', () => {
		expect(
			findRedirected([
				{
					requestedUrl: 'http://localhost:4191/pipelines',
					finalDisplayedUrl: 'http://localhost:4191/pipelines',
				},
			]),
		).toEqual([]);
	});

	test('flags a page that bounced to the login page', () => {
		expect(
			findRedirected([
				{
					requestedUrl: 'http://localhost:4191/pipelines',
					finalDisplayedUrl: 'http://localhost:4191/login',
				},
				{
					requestedUrl: 'http://localhost:4191/login',
					finalDisplayedUrl: 'http://localhost:4191/login',
				},
			]),
		).toEqual([
			'http://localhost:4191/pipelines ended on http://localhost:4191/login',
		]);
	});

	test('ignores a trailing slash', () => {
		expect(
			findRedirected([
				{
					requestedUrl: 'http://localhost:4191/',
					finalDisplayedUrl: 'http://localhost:4191',
				},
			]),
		).toEqual([]);
	});
});
