import { beforeEach, describe, expect, test } from 'bun:test';
import {
	DEFAULT_HEALTH_FILTER,
	DEFAULT_REPO_SELECTOR,
	DEFAULT_SORT_BY,
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

beforeEach(() => {
	resetFilters();
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

		expect(getHealthFilter()).toBe(DEFAULT_HEALTH_FILTER);
		expect(getRepoSelector()).toBe(DEFAULT_REPO_SELECTOR);
		expect(getSortBy()).toBe(DEFAULT_SORT_BY);
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

		expect(getHealthFilter()).toBe(DEFAULT_HEALTH_FILTER);
		expect(getRepoSelector()).toBe(DEFAULT_REPO_SELECTOR);
		expect(getSortBy()).toBe(DEFAULT_SORT_BY);
	});
});
