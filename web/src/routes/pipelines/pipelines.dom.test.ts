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
import { initPipelinesFilters } from '#lib/pipelinesFilters.svelte.js';
import type { ServerSettings, SettingsValues } from '#lib/settingsSync.js';
import { removeDom } from '../../../test/dom.js';
import { fakeNetwork, json, settle } from '../../../test/page.js';

// The Pipelines page against a fake network: what it says when a filter matches
// nothing and when nothing has been ingested, and how "Reset filters" follows the
// defaults the server sent. Which pipelines a filter keeps is the server's; the
// fake applies the health filter the way the server does so the page's own
// wiring is what is under test.

mock.module('$app/state', () => ({
	page: {
		data: { hasRepos: true },
		url: new URL('http://localhost/pipelines'),
	},
}));

const { default: Page } = await import('./+page.svelte');

const serverDefaults: SettingsValues = {
	theme: 'system',
	forgeFilter: 'all',
	pipelinesHealthFilter: 'unhealthy',
	pipelinesRepoSelector: 'all',
	pipelinesSortOrder: 'name',
	telemetryWindow: '7d',
};

function settings(): ServerSettings {
	return { ...serverDefaults, defaults: serverDefaults };
}

interface Fixture {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
}

let network: ReturnType<typeof fakeNetwork>;
let fixtures: Fixture[];
let app: ReturnType<typeof mount> | undefined;

beforeEach(() => {
	localStorage.clear();
	initPipelinesFilters(settings());
	fixtures = [];

	network = fakeNetwork((url) => {
		if (url.pathname === '/api/pipelines') {
			const health = url.searchParams.get('health');

			return json({
				pipelines: fixtures.filter((p) => !health || p.healthStatus === health),
				hasMore: false,
			});
		}
		if (url.pathname === '/api/repos') {
			return json({
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			});
		}
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

// The forge filter has an "All" of its own, so look inside the health group.
function healthRadio(name: string): HTMLElement {
	const group = document.querySelector(
		'[aria-label="Filter by health status"]',
	);
	const found = [
		...(group?.querySelectorAll<HTMLElement>('[role="radio"]') ?? []),
	].find((el) => el.textContent?.trim() === name);
	if (!found) throw new Error(`no health radio named ${name}`);

	return found;
}

function resetButton(): HTMLButtonElement {
	const found = [...document.querySelectorAll('button')].find(
		(el) => el.textContent?.trim() === 'Reset filters',
	);
	if (!found) throw new Error('no Reset filters button');

	return found;
}

describe('Pipelines page', () => {
	test('a filter that matches nothing says so and keeps the controls', async () => {
		// Every pipeline is healthy, and the filter defaults to Unhealthy.
		fixtures = [
			{ id: 'ci', repoId: 'r1', name: 'CI', healthStatus: 'healthy' },
		];

		await open();

		expect(document.body.textContent).toContain(
			'No pipelines match the selected filters.',
		);
		expect(document.body.textContent).not.toContain(
			'No pipeline runs ingested yet',
		);
		expect(
			document.querySelector('[aria-label="Filter by health status"]'),
		).not.toBeNull();

		healthRadio('All').click();
		await settle();

		expect(document.querySelector('h3')?.textContent).toBe('CI');
	});

	test('with no pipelines at all, it still says nothing has been ingested', async () => {
		await open();

		expect(document.body.textContent).toContain(
			'No pipeline runs ingested yet',
		);
		expect(document.body.textContent).not.toContain(
			'No pipelines match the selected filters.',
		);
	});

	test('"Reset filters" follows the server\'s defaults, with no copy in the page', async () => {
		fixtures = [
			{ id: 'ci', repoId: 'r1', name: 'CI', healthStatus: 'unhealthy' },
		];
		await open();

		// A fresh account is on the server's defaults, whatever they are.
		expect(resetButton().disabled).toBe(true);

		healthRadio('All').click();
		await settle();
		expect(resetButton().disabled).toBe(false);

		resetButton().click();
		await settle();

		expect(resetButton().disabled).toBe(true);
		expect(healthRadio('Unhealthy').getAttribute('aria-checked')).toBe('true');
	});
});
