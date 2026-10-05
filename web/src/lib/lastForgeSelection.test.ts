import { afterEach, beforeEach, describe, expect, test } from 'bun:test';
import {
	getRememberedForgeSelection,
	purgeRememberedTokens,
	rememberForgeSelection,
} from './lastForgeSelection.js';

const realStorage = Object.getOwnPropertyDescriptor(globalThis, 'localStorage');

let store: Map<string, string>;
let removed: string[];

function useStorage(storage: Partial<Storage>): void {
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
	removeItem: () => {
		throw new Error('storage unavailable');
	},
	key: () => {
		throw new Error('storage unavailable');
	},
	get length(): number {
		throw new Error('storage unavailable');
	},
};

beforeEach(() => {
	store = new Map();
	removed = [];

	useStorage({
		getItem: (key: string) => store.get(key) ?? null,
		setItem: (key: string, value: string) => void store.set(key, value),
		removeItem: (key: string) => {
			removed.push(key);
			store.delete(key);
		},
		// Like the real thing it has nothing past the last index; asking
		// anyway is a bug, so it throws rather than answering null.
		key: (index: number) => {
			const keys = [...store.keys()];
			if (index >= keys.length) throw new RangeError('no such key');

			return keys[index];
		},
		get length() {
			return store.size;
		},
	});
});

afterEach(() => {
	if (realStorage) {
		Object.defineProperty(globalThis, 'localStorage', realStorage);
	} else {
		Reflect.deleteProperty(globalThis, 'localStorage');
	}
});

describe('the last forge selection', () => {
	test('is nothing until one is remembered', () => {
		expect(getRememberedForgeSelection()).toBeNull();
	});

	test('round-trips a GitHub selection', () => {
		rememberForgeSelection('github', undefined);

		expect(getRememberedForgeSelection()).toEqual({ forge: 'github' });
		expect(store.has('lastForgeSelection')).toBe(true);
	});

	test('round-trips a Forgejo selection with its instance', () => {
		rememberForgeSelection('forgejo', 'https://forgejo.example.com');

		expect(getRememberedForgeSelection()).toEqual({
			forge: 'forgejo',
			instanceUrl: 'https://forgejo.example.com',
		});
	});

	test('drops an instance URL that is not Forgejo, and an empty one', () => {
		rememberForgeSelection('github', 'https://ignored');
		expect(getRememberedForgeSelection()).toEqual({ forge: 'github' });

		rememberForgeSelection('forgejo', '');
		expect(getRememberedForgeSelection()).toEqual({ forge: 'forgejo' });
	});

	test('survives unavailable storage', () => {
		useStorage(brokenStorage);

		rememberForgeSelection('github', undefined);

		expect(getRememberedForgeSelection()).toBeNull();
	});

	test('is nothing when the stored value is not JSON', () => {
		store.set('lastForgeSelection', '{nope');

		expect(getRememberedForgeSelection()).toBeNull();
	});
});

describe('purgeRememberedTokens', () => {
	test('removes every token the browser remembered, and nothing else', () => {
		store.set('rememberedToken:github:', 'ghp_secret');
		store.set('rememberedToken:forgejo:https://f.example', 'fj_secret');
		store.set('rememberedToken:lastSelection', '{"forge":"github"}');
		store.set('lastForgeSelection', '{"forge":"github"}');
		store.set('telemetryWindow', '7d');

		purgeRememberedTokens();

		expect([...store.keys()].sort()).toEqual([
			'lastForgeSelection',
			'telemetryWindow',
		]);
		expect(removed.sort()).toEqual([
			'rememberedToken:forgejo:https://f.example',
			'rememberedToken:github:',
			'rememberedToken:lastSelection',
		]);
	});

	test('skips an index the storage cannot name', () => {
		store.set('rememberedToken:github:', 'ghp_secret');
		const keys = ['rememberedToken:github:', null];
		useStorage({
			removeItem: (key: string) => void store.delete(key),
			key: (index: number) => keys[index] ?? null,
			get length() {
				return keys.length;
			},
		});

		purgeRememberedTokens();

		expect(store.size).toBe(0);
	});

	test('does nothing when storage is unavailable', () => {
		useStorage(brokenStorage);

		expect(() => purgeRememberedTokens()).not.toThrow();
	});
});
