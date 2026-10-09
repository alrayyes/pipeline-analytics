import { describe, expect, test } from 'bun:test';
import { preload } from './hooks.server.js';

describe('preload', () => {
	test('preloads the latin Inter subset', () => {
		expect(
			preload({
				type: 'font',
				path: '/_app/immutable/assets/inter-latin-wght-normal.Dx4kXJAl.woff2',
			}),
		).toBe(true);
	});

	test('leaves the other Inter subsets to load on demand', () => {
		for (const subset of ['latin-ext', 'cyrillic', 'greek', 'vietnamese']) {
			expect(
				preload({
					type: 'font',
					path: `/_app/immutable/assets/inter-${subset}-wght-normal.Dx4kXJAl.woff2`,
				}),
			).toBe(false);
		}
	});

	test("keeps SvelteKit's own default for scripts and styles", () => {
		expect(
			preload({ type: 'js', path: '/_app/immutable/entry/start.js' }),
		).toBe(true);
		expect(preload({ type: 'css', path: '/_app/immutable/assets/0.css' })).toBe(
			true,
		);
	});
});
