import { beforeEach, describe, expect, test } from 'bun:test';
import {
	getOffset,
	getStatus,
	nextPage,
	previousPage,
	RUNS_PAGE_SIZE,
	resetRunsFilters,
	setStatus,
} from './runsFilters.svelte.js';

beforeEach(() => {
	resetRunsFilters();
});

describe('defaults', () => {
	test('starts on every run, first page', () => {
		expect(getStatus()).toBe('all');
		expect(getOffset()).toBe(0);
	});

	test('a page is 20 runs, like the Pipelines and Repos pages', () => {
		expect(RUNS_PAGE_SIZE).toBe(20);
	});
});

describe('setStatus', () => {
	test('a set status round-trips through its getter', () => {
		setStatus('failed');

		expect(getStatus()).toBe('failed');
	});

	test('changing the status returns to the first page', () => {
		nextPage();
		nextPage();
		expect(getOffset()).toBe(2 * RUNS_PAGE_SIZE);

		setStatus('running');

		expect(getOffset()).toBe(0);
	});

	test('choosing the status already selected keeps the page', () => {
		setStatus('failed');
		nextPage();

		setStatus('failed');

		expect(getOffset()).toBe(RUNS_PAGE_SIZE);
	});
});

describe('paging', () => {
	test('next advances by one page', () => {
		nextPage();

		expect(getOffset()).toBe(RUNS_PAGE_SIZE);
	});

	test('previous steps back one page', () => {
		nextPage();
		nextPage();

		previousPage();

		expect(getOffset()).toBe(RUNS_PAGE_SIZE);
	});

	test('previous never goes below the first page', () => {
		previousPage();

		expect(getOffset()).toBe(0);
	});
});

describe('resetRunsFilters', () => {
	test('restores the status and the page', () => {
		setStatus('success');
		nextPage();

		resetRunsFilters();

		expect(getStatus()).toBe('all');
		expect(getOffset()).toBe(0);
	});
});
