import { afterEach, beforeEach, describe, expect, test } from 'bun:test';
import {
	getHealthFilter,
	getRepoSelector,
	getSortBy,
	initPipelinesFilters,
	isAtDefaults,
	resetFilters,
	setHealthFilter,
	setRepoSelector,
	setSortBy,
} from './pipelinesFilters.svelte.js';

const realFetch = globalThis.fetch;
const realStorage = Object.getOwnPropertyDescriptor(globalThis, 'localStorage');

let patches: unknown[];
let store: Map<string, string>;

function useStorage(storage: Pick<Storage, 'getItem' | 'setItem'>): void {
	Object.defineProperty(globalThis, 'localStorage', {
		configurable: true,
		value: storage,
	});
}

const brokenStorage = {
	getItem: () => {
		throw new Error('storage unavailable');
	},
	setItem: () => {
		throw new Error('storage unavailable');
	},
};

beforeEach(() => {
	patches = [];
	store = new Map();

	globalThis.fetch = ((_url: unknown, init?: RequestInit) => {
		patches.push(JSON.parse(String(init?.body)));
		return Promise.resolve(new Response('{}'));
	}) as typeof fetch;

	useStorage({
		getItem: (key: string) => store.get(key) ?? null,
		setItem: (key: string, value: string) => void store.set(key, value),
	});
});

afterEach(() => {
	globalThis.fetch = realFetch;

	if (realStorage) {
		Object.defineProperty(globalThis, 'localStorage', realStorage);
	} else {
		Reflect.deleteProperty(globalThis, 'localStorage');
	}
});

beforeEach(() => {
	resetFilters();
	patches = [];
	store.clear();
});

describe('getters/setters', () => {
	test('a set value round-trips through its getter', () => {
		setHealthFilter('healthy');
		expect(getHealthFilter()).toBe('healthy');

		setRepoSelector('repo-1');
		expect(getRepoSelector()).toBe('repo-1');

		setSortBy('lastRun');
		expect(getSortBy()).toBe('lastRun');
	});
});

describe('isAtDefaults', () => {
	test('true when every filter is still its default', () => {
		expect(isAtDefaults()).toBe(true);
	});

	test('false once any single filter moves off its default', () => {
		setHealthFilter('all');
		expect(isAtDefaults()).toBe(false);
	});
});

describe('resetFilters', () => {
	test('restores every filter to its documented default', () => {
		setHealthFilter('all');
		setRepoSelector('repo-1');
		setSortBy('lastRun');

		resetFilters();

		expect(getHealthFilter()).toBe('unhealthy');
		expect(getRepoSelector()).toBe('all');
		expect(getSortBy()).toBe('name');
	});
});

describe('initPipelinesFilters', () => {
	test('server-supplied values win over the cache and the default', () => {
		initPipelinesFilters({
			pipelinesHealthFilter: 'all',
			pipelinesRepoSelector: 'repo-2',
			pipelinesSortOrder: 'lastRun',
		});

		expect(getHealthFilter()).toBe('all');
		expect(getRepoSelector()).toBe('repo-2');
		expect(getSortBy()).toBe('lastRun');
	});

	test('falls back to the documented default when no server value is given', () => {
		initPipelinesFilters(undefined);

		expect(getHealthFilter()).toBe('unhealthy');
		expect(getRepoSelector()).toBe('all');
		expect(getSortBy()).toBe('name');
	});
});

describe('isAtDefaults, one filter at a time', () => {
	test('a repo selector off its default is not at defaults', () => {
		setRepoSelector('repo-1');
		expect(isAtDefaults()).toBe(false);
	});

	test('a sort order off its default is not at defaults', () => {
		setSortBy('lastRun');
		expect(isAtDefaults()).toBe(false);
	});
});

describe('persistence', () => {
	test('each setter caches under its own key and patches only its own setting', async () => {
		setHealthFilter('healthy');
		setRepoSelector('repo-1');
		setSortBy('lastRun');
		await Promise.resolve();

		expect(Object.fromEntries(store)).toEqual({
			pipelinesHealthFilter: 'healthy',
			pipelinesRepoSelector: 'repo-1',
			pipelinesSortOrder: 'lastRun',
		});
		expect(patches).toEqual([
			{ pipelinesHealthFilter: 'healthy' },
			{ pipelinesRepoSelector: 'repo-1' },
			{ pipelinesSortOrder: 'lastRun' },
		]);
	});

	test('reset caches the defaults and clears all three on the server', async () => {
		setHealthFilter('all');
		setRepoSelector('repo-1');
		setSortBy('lastRun');
		patches = [];

		resetFilters();
		await Promise.resolve();

		expect(Object.fromEntries(store)).toEqual({
			pipelinesHealthFilter: 'unhealthy',
			pipelinesRepoSelector: 'all',
			pipelinesSortOrder: 'name',
		});
		expect(patches).toEqual([
			{
				pipelinesHealthFilter: null,
				pipelinesRepoSelector: null,
				pipelinesSortOrder: null,
			},
		]);
	});

	test('setters still work when storage is unavailable', () => {
		useStorage(brokenStorage);

		setHealthFilter('healthy');

		expect(getHealthFilter()).toBe('healthy');
	});
});

describe('initPipelinesFilters from the cache', () => {
	test('reads each cached value under its own key', () => {
		store.set('pipelinesHealthFilter', 'healthy');
		store.set('pipelinesRepoSelector', 'repo-3');
		store.set('pipelinesSortOrder', 'lastRun');

		initPipelinesFilters(undefined);

		expect(getHealthFilter()).toBe('healthy');
		expect(getRepoSelector()).toBe('repo-3');
		expect(getSortBy()).toBe('lastRun');
	});

	test('writes the resolved values back to the cache', () => {
		initPipelinesFilters({
			pipelinesHealthFilter: 'all',
			pipelinesRepoSelector: 'repo-2',
			pipelinesSortOrder: 'lastRun',
		});

		expect(Object.fromEntries(store)).toEqual({
			pipelinesHealthFilter: 'all',
			pipelinesRepoSelector: 'repo-2',
			pipelinesSortOrder: 'lastRun',
		});
	});

	test('falls back to the defaults when storage is unavailable', () => {
		useStorage(brokenStorage);

		initPipelinesFilters(undefined);

		expect(getHealthFilter()).toBe('unhealthy');
		expect(getRepoSelector()).toBe('all');
		expect(getSortBy()).toBe('name');
	});
});
