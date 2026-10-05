import { expect, test } from 'bun:test';
import { getForgeFilter } from './forgeFilter.svelte.js';

// Its own file: the other tests set the filter in beforeEach, so only a
// file that reads it first can see the value the module starts with.
test('starts on "all"', () => {
	expect(getForgeFilter()).toBe('all');
});
