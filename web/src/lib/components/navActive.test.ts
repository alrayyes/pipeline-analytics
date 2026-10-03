import { describe, expect, test } from 'bun:test';
import { isActivePath } from './navActive.js';

describe('isActivePath', () => {
	test('the root is active only on exactly the root', () => {
		expect(isActivePath('/', '/')).toBe(true);
		expect(isActivePath('/runs', '/')).toBe(false);
		expect(isActivePath('/pipelines/abc', '/')).toBe(false);
	});

	test('any other link is active on its own path', () => {
		expect(isActivePath('/runs', '/runs')).toBe(true);
	});

	test('and on a path beneath it, so a detail page keeps its section lit', () => {
		expect(isActivePath('/pipelines/abc', '/pipelines')).toBe(true);
		expect(isActivePath('/pipelines/abc/flaky-runs', '/pipelines')).toBe(true);
	});

	test('but not on an unrelated path', () => {
		expect(isActivePath('/failures', '/runs')).toBe(false);
		expect(isActivePath('/', '/runs')).toBe(false);
	});
});
