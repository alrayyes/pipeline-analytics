import { beforeEach, describe, expect, test } from 'bun:test';
import {
	getForgeFilter,
	initForgeFilter,
	setForgeFilter,
} from './forgeFilter.svelte.js';

beforeEach(() => {
	setForgeFilter('all');
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
