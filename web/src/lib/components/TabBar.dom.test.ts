import '../../../test/dom.js';
import {
	afterAll,
	afterEach,
	beforeEach,
	describe,
	expect,
	mock,
	test,
} from 'bun:test';
import { mount, unmount } from 'svelte';
import { removeDom } from '../../../test/dom.js';

// The phone tab bar: which tabs are links and which are disabled like the top
// nav, and which one is marked as the current page. That it sits at the bottom of
// a phone and not on a desktop is layout, and stays in e2e.

const current = { pathname: '/' };

mock.module('$app/state', () => ({
	page: {
		get url() {
			return new URL(`http://localhost${current.pathname}`);
		},
	},
}));

const { default: TabBar } = await import('./TabBar.svelte');

let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	current.pathname = '/';
});

afterEach(() => {
	if (app) unmount(app);
	app = undefined;
	document.body.innerHTML = '';
});

afterAll(async () => {
	await removeDom();
});

function open(hasRepos: boolean): HTMLElement {
	app = mount(TabBar, { target: document.body, props: { hasRepos } });

	const nav = document.querySelector<HTMLElement>(
		'nav[aria-label="Main views"]',
	);
	if (!nav) throw new Error('no tab bar');

	return nav;
}

function tab(nav: HTMLElement, label: string): HTMLElement {
	const found = [...nav.querySelectorAll<HTMLElement>('a, [role="link"]')].find(
		(el) => el.textContent?.trim() === label,
	);
	if (!found) throw new Error(`no tab named ${label}`);

	return found;
}

describe('TabBar', () => {
	test('with no repository registered, the tabs that need one are disabled like the top nav', () => {
		const nav = open(false);

		expect(tab(nav, 'Overview').getAttribute('href')).toBe('/');
		for (const label of ['Pipelines', 'Runs', 'Failures']) {
			expect(tab(nav, label).getAttribute('aria-disabled')).toBe('true');
			expect(tab(nav, label).hasAttribute('href')).toBe(false);
		}
	});

	test('with a repository every tab is a link to its view', () => {
		const nav = open(true);

		expect(
			['Overview', 'Pipelines', 'Runs', 'Failures'].map((label) =>
				tab(nav, label).getAttribute('href'),
			),
		).toEqual(['/', '/pipelines', '/runs', '/failures']);
		expect(nav.querySelector('[aria-disabled]')).toBeNull();
	});

	test('marks the current view and no other', () => {
		current.pathname = '/runs';
		const nav = open(true);

		expect(tab(nav, 'Runs').getAttribute('aria-current')).toBe('page');
		expect(nav.querySelectorAll('[aria-current]')).toHaveLength(1);
	});
});
