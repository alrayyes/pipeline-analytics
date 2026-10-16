import '../../test/dom.js';
import { afterAll, afterEach, describe, expect, mock, test } from 'bun:test';
import { createRawSnippet, mount, unmount } from 'svelte';
import { removeDom } from '../../test/dom.js';
import { fakeNetwork, json, settle } from '../../test/page.js';

// The root layout's wiring that no page test reaches: the WebMCP tools are
// registered with the router's invalidateAll, which is how update_settings makes
// the open page show a change an agent made.

const invalidateAll = mock(() => Promise.resolve());
const registerWebMCPTools = mock(
	(_context?: unknown, _options?: { refreshPage?: () => Promise<void> }) => {},
);

mock.module('$app/navigation', () => ({
	invalidateAll,
	goto: () => Promise.resolve(),
	afterNavigate: () => {},
	beforeNavigate: () => {},
}));
mock.module('$app/state', () => ({
	page: { url: new URL('http://localhost/'), data: {} },
}));
mock.module('#lib/webmcpTools.js', () => ({ registerWebMCPTools }));
mock.module('#lib/serviceWorker.js', () => ({
	registerServiceWorker: () => {},
}));

const { default: Layout } = await import('./+layout.svelte');

const network = fakeNetwork((url) => {
	if (url.pathname === '/api/version') return json({ version: 'dev' });
});

let app: ReturnType<typeof mount> | undefined;

afterEach(() => {
	if (app) unmount(app);
	app = undefined;
	document.body.innerHTML = '';
});

afterAll(async () => {
	network.restore();
	await removeDom();
});

describe('root layout', () => {
	test("registers the WebMCP tools with the router's invalidateAll", async () => {
		app = mount(Layout, {
			target: document.body,
			props: {
				data: { hasRepos: true, settings: null },
				children: createRawSnippet(() => ({ render: () => '<main></main>' })),
			},
		});
		await settle();

		expect(registerWebMCPTools).toHaveBeenCalledTimes(1);
		expect(registerWebMCPTools.mock.calls[0]?.[1]?.refreshPage).toBe(
			invalidateAll,
		);
	});
});
