import { describe, expect, test } from 'bun:test';
import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { Glob } from 'bun';
import {
	END_MARKER,
	EXCLUDED_ROUTES,
	EXTRA_SCREENSHOTS,
	renderScreenshotSection,
	replaceScreenshotSection,
	routeOf,
	SCREENSHOTS,
	START_MARKER,
} from '../../scripts/screenshot-routes.js';

// Stryker mutates a copy of web/ alone, with no README.md or docs/ beside it,
// and this checks those, not the code it mutates; so it skips there and only
// there (not on a missing README, which must fail).
const inMutationSandbox = import.meta.dir.includes('.stryker-tmp');

const ROUTES_DIR = fileURLToPath(new URL('../routes/', import.meta.url));
const DOCS_DIR = fileURLToPath(new URL('../../../docs/', import.meta.url));
const README = inMutationSandbox
	? ''
	: readFileSync(
			fileURLToPath(new URL('../../../README.md', import.meta.url)),
			'utf8',
		);
const CAPTURE_SCRIPT = readFileSync(
	fileURLToPath(
		new URL('../../scripts/capture-screenshots.mjs', import.meta.url),
	),
	'utf8',
);

const routes = [...new Glob('**/+page.svelte').scanSync(ROUTES_DIR)]
	.map(routeOf)
	.sort();

describe('routeOf', () => {
	test('turns a page file into its route', () => {
		expect(routeOf('+page.svelte')).toBe('/');
		expect(routeOf('failures/+page.svelte')).toBe('/failures');
		expect(routeOf('pipelines/[id]/flaky-runs/+page.svelte')).toBe(
			'/pipelines/[id]/flaky-runs',
		);
	});
});

describe.skipIf(inMutationSandbox)('README screenshots', () => {
	test('every page under src/routes has one, except the footer pages', () => {
		const covered = new Set([...Object.keys(SCREENSHOTS), ...EXCLUDED_ROUTES]);

		expect(routes.filter((route) => !covered.has(route))).toEqual([]);
	});

	test('the footer pages have none', () => {
		expect(EXCLUDED_ROUTES).toEqual(['/legal', '/releases']);
		for (const route of EXCLUDED_ROUTES) {
			expect(SCREENSHOTS).not.toHaveProperty(route);
		}
	});

	test('no entry names a page that no longer exists', () => {
		expect(
			Object.keys(SCREENSHOTS).filter((route) => !routes.includes(route)),
		).toEqual([]);
	});

	for (const [route, { file, alt }] of Object.entries(SCREENSHOTS)) {
		test(`${route}: ${file} exists, is captured, and has alt text`, () => {
			expect(existsSync(`${DOCS_DIR}${file}`)).toBe(true);
			expect(CAPTURE_SCRIPT).toContain(`'${file}'`);
			expect(alt.length).toBeGreaterThanOrEqual(10);
		});
	}

	for (const { file } of EXTRA_SCREENSHOTS) {
		test(`${file} exists and is captured`, () => {
			expect(existsSync(`${DOCS_DIR}${file}`)).toBe(true);
			expect(CAPTURE_SCRIPT).toContain(`'${file}'`);
		});
	}

	test('the README section is the generated one (bun scripts/update-readme-screenshots.ts)', () => {
		const start = README.indexOf(START_MARKER);
		const end = README.indexOf(END_MARKER) + END_MARKER.length;

		expect(README.slice(start, end)).toBe(renderScreenshotSection());
	});
});

describe('replaceScreenshotSection', () => {
	test('swaps what is between the markers and keeps the rest', () => {
		const out = replaceScreenshotSection(
			`before\n${START_MARKER}\nold\n${END_MARKER}\nafter`,
		);

		expect(out).toBe(`before\n${renderScreenshotSection()}\nafter`);
	});

	test('refuses a README with no markers', () => {
		expect(() => replaceScreenshotSection('nothing here')).toThrow(
			'README needs',
		);
	});
});
