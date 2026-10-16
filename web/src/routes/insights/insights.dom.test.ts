import '../../../test/dom.js';
import {
	afterAll,
	afterEach,
	beforeEach,
	describe,
	expect,
	test,
} from 'bun:test';
import { mount, unmount } from 'svelte';
import { removeDom } from '../../../test/dom.js';
import { fakeNetwork, json, settle } from '../../../test/page.js';
import Page from './+page.svelte';

// The Insights page against a fake network: what it says when the rate-limit
// request is refused and when the server cannot be reached. The figures are the
// server's; the wording of each failure is the page's.

let network: ReturnType<typeof fakeNetwork>;
let respond: () => Response;
let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	respond = () => json([]);
	network = fakeNetwork((url) => {
		if (url.pathname === '/api/insights/github-rate-limit') return respond();
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

async function open(): Promise<void> {
	app = mount(Page, { target: document.body });
	await settle();
}

describe('Insights page', () => {
	test('says so when the server refuses the rate-limit request', async () => {
		respond = () => json({}, 500);

		await open();

		expect(document.body.textContent).toContain(
			'Could not load GitHub API usage.',
		);
	});

	test('says so when the server cannot be reached', async () => {
		network.restore();
		globalThis.fetch = (() =>
			Promise.reject(new TypeError('offline'))) as typeof fetch;

		await open();

		expect(document.body.textContent).toContain('Could not reach the server.');
	});
});
