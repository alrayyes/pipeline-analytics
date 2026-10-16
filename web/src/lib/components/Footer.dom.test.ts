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
import { fakeNetwork, json, settle } from '../../../test/page.js';

// The footer against a fake /api/version: what it says for a dev build and a
// release, where each link goes and which ones leave the app in a new tab. How it
// wraps on a phone, and its axe scan, need a real browser and stay in e2e.

const current = { pathname: '/login' };

mock.module('$app/state', () => ({
	page: {
		get url() {
			return new URL(`http://localhost${current.pathname}`);
		},
	},
}));

const { default: Footer } = await import('./Footer.svelte');

let network: ReturnType<typeof fakeNetwork>;
let version: string;
let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	current.pathname = '/login';
	version = 'dev';
	network = fakeNetwork((url) => {
		if (url.pathname === '/api/version') return json({ version });
	});
});

afterEach(() => {
	if (app) unmount(app);
	app = undefined;
	document.body.innerHTML = '';
});

afterAll(async () => {
	network.restore();
	await removeDom();
});

async function open(): Promise<HTMLElement> {
	app = mount(Footer, { target: document.body });
	await settle();

	const footer = document.querySelector('footer');
	if (!footer) throw new Error('no footer');

	return footer;
}

function link(
	footer: HTMLElement,
	text: string,
): HTMLAnchorElement | undefined {
	return [...footer.querySelectorAll('a')].find((a) =>
		a.textContent?.trim().startsWith(text),
	);
}

describe('Footer', () => {
	test('a dev build says so and links nowhere', async () => {
		const footer = await open();

		expect(footer.textContent).toContain('dev build');
		expect(link(footer, 'dev')).toBeUndefined();
		expect(link(footer, 'Release history')).toBeUndefined();
	});

	test('a release version links to the release history, with no second link to it', async () => {
		version = '0.56.1';

		const footer = await open();

		expect(link(footer, '0.56.1')?.getAttribute('href')).toBe('/releases');
		expect(link(footer, 'Release history')).toBeUndefined();
		// The privacy link beside it is untouched.
		expect(link(footer, 'Privacy')).toBeDefined();
	});

	test('on the release page the version is not a link to itself', async () => {
		version = '0.56.1';
		current.pathname = '/releases';

		const footer = await open();

		expect(footer.textContent).toContain('0.56.1');
		expect(link(footer, '0.56.1')).toBeUndefined();
	});

	test('on the privacy page there is no link to itself', async () => {
		current.pathname = '/legal';

		const footer = await open();

		expect(link(footer, 'Privacy')).toBeUndefined();
	});

	test('the GitHub and license links point at the repo, open in a new tab and say so', async () => {
		version = '0.56.1';
		const footer = await open();

		const github = link(footer, 'GitHub');
		const license = link(footer, 'AGPL-3.0');
		expect(github?.getAttribute('href')).toBe(
			'https://github.com/alrayyes/pipeline-analytics',
		);
		expect(license?.getAttribute('href')).toBe(
			'https://github.com/alrayyes/pipeline-analytics/blob/main/LICENSE',
		);
		for (const a of [github, license]) {
			expect(a?.getAttribute('target')).toBe('_blank');
			expect(a?.getAttribute('rel')).toBe('noopener noreferrer');
			expect(a?.textContent).toContain('(opens in a new tab)');
		}
		// The mark sits beside its own label, so it is decorative.
		expect(footer.querySelectorAll('svg[aria-hidden="true"]')).toHaveLength(1);

		// Same-tab links have no target.
		expect(link(footer, '0.56.1')?.hasAttribute('target')).toBe(false);
		expect(link(footer, 'Privacy')?.hasAttribute('target')).toBe(false);
	});
});
