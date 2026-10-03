import { afterEach, beforeEach, describe, expect, test } from 'bun:test';
import {
	getTelemetryWindow,
	initTelemetryWindow,
	setTelemetryWindow,
} from './telemetryWindow.svelte.js';

const realFetch = globalThis.fetch;
const realStorage = Object.getOwnPropertyDescriptor(globalThis, 'localStorage');

let patches: unknown[];
let store: Map<string, string>;

beforeEach(() => {
	patches = [];
	store = new Map();

	globalThis.fetch = ((_url: unknown, init?: RequestInit) => {
		patches.push(JSON.parse(String(init?.body)));
		return Promise.resolve(new Response('{}'));
	}) as typeof fetch;

	Object.defineProperty(globalThis, 'localStorage', {
		configurable: true,
		value: {
			getItem: (key: string) => store.get(key) ?? null,
			setItem: (key: string, value: string) => void store.set(key, value),
		},
	});

	initTelemetryWindow(undefined);
});

afterEach(() => {
	globalThis.fetch = realFetch;

	if (realStorage) {
		Object.defineProperty(globalThis, 'localStorage', realStorage);
	} else {
		Reflect.deleteProperty(globalThis, 'localStorage');
	}
});

describe('with nothing saved', () => {
	test('there is no window: the server chooses, and the page shows what it used', () => {
		expect(getTelemetryWindow()).toBeNull();
	});
});

describe('setTelemetryWindow', () => {
	test('a set value round-trips through its getter', () => {
		setTelemetryWindow('30d');
		expect(getTelemetryWindow()).toBe('30d');

		setTelemetryWindow('24h');
		expect(getTelemetryWindow()).toBe('24h');
	});

	test('sends the choice to the server as a settings patch', () => {
		setTelemetryWindow('30d');

		expect(patches).toEqual([{ telemetryWindow: '30d' }]);
	});

	test('caches the choice so a reload paints it before the server answers', () => {
		setTelemetryWindow('24h');

		expect(store.get('telemetryWindow')).toBe('24h');
	});
});

describe('initTelemetryWindow', () => {
	test('a server-supplied window wins', () => {
		store.set('telemetryWindow', '24h');

		initTelemetryWindow('30d');

		expect(getTelemetryWindow()).toBe('30d');
	});

	test('falls back to the cached window when the server gave none', () => {
		store.set('telemetryWindow', '24h');

		initTelemetryWindow(undefined);

		expect(getTelemetryWindow()).toBe('24h');
	});

	test('is empty when nothing is cached either: the frontend holds no default', () => {
		initTelemetryWindow(undefined);

		expect(getTelemetryWindow()).toBeNull();
	});

	test('ignores a cached value outside the documented set', () => {
		store.set('telemetryWindow', '12');

		initTelemetryWindow(undefined);

		expect(getTelemetryWindow()).toBeNull();
	});

	test('caches nothing when there is nothing to cache', () => {
		initTelemetryWindow(undefined);

		expect(store.has('telemetryWindow')).toBe(false);
	});

	test('does not send a patch: it only reconciles with what the server already has', () => {
		initTelemetryWindow('30d');

		expect(patches).toEqual([]);
	});
});
