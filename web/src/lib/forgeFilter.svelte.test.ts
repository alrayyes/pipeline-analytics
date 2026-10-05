import { afterEach, beforeEach, describe, expect, test } from 'bun:test';
import {
	getForgeFilter,
	initForgeFilter,
	setForgeFilter,
} from './forgeFilter.svelte.js';

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
	initForgeFilter('all');
	patches = [];
	store.clear();
});

test('a set value round-trips through its getter', () => {
	setForgeFilter('github');
	expect(getForgeFilter()).toBe('github');

	setForgeFilter('forgejo');
	expect(getForgeFilter()).toBe('forgejo');
});

describe('initForgeFilter', () => {
	test('a server-supplied filter wins', () => {
		initForgeFilter('github');
		expect(getForgeFilter()).toBe('github');
	});

	test('falls back to "all" when no server filter and nothing cached', () => {
		initForgeFilter(undefined);
		expect(getForgeFilter()).toBe('all');
	});
});

describe('setForgeFilter', () => {
	test('caches the value under "forgeFilter" and patches the server', async () => {
		setForgeFilter('forgejo');
		await Promise.resolve();

		expect(store.get('forgeFilter')).toBe('forgejo');
		expect(patches).toEqual([{ forgeFilter: 'forgejo' }]);
	});

	test('still sets the value when storage is unavailable', () => {
		useStorage(brokenStorage);

		setForgeFilter('github');

		expect(getForgeFilter()).toBe('github');
	});
});

describe('initForgeFilter from the cache', () => {
	test.each(['github', 'forgejo'] as const)('reads a cached "%s"', (cached) => {
		store.set('forgeFilter', cached);

		initForgeFilter(undefined);

		expect(getForgeFilter()).toBe(cached);
	});

	test('ignores a cached value that is not a forge filter', () => {
		store.set('forgeFilter', 'bitbucket');

		initForgeFilter(undefined);

		expect(getForgeFilter()).toBe('all');
	});

	test('writes the resolved value back to the cache', () => {
		initForgeFilter('forgejo');

		expect(store.get('forgeFilter')).toBe('forgejo');
	});

	test('falls back to "all" when storage is unavailable', () => {
		useStorage(brokenStorage);

		initForgeFilter(undefined);

		expect(getForgeFilter()).toBe('all');
	});
});
