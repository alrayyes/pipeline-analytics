import { expect, test } from 'bun:test';
import {
	getHealthFilter,
	getRepoSelector,
	getSortBy,
} from './pipelinesFilters.svelte.js';

// Its own file, for the reason forgeFilter.initial.test.ts gives.
test('start on the documented defaults', () => {
	expect(getHealthFilter()).toBe('unhealthy');
	expect(getRepoSelector()).toBe('all');
	expect(getSortBy()).toBe('name');
});
