import { expect, test } from 'bun:test';
import {
	getHealthFilter,
	getRepoSelector,
	getSortBy,
} from './pipelinesFilters.svelte.js';

// Its own file, for the reason forgeFilter.initial.test.ts gives.
test('start unfiltered until the server answers', () => {
	expect(getHealthFilter()).toBe('all');
	expect(getRepoSelector()).toBe('all');
	expect(getSortBy()).toBe('name');
});
