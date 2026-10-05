import { describe, expect, test } from 'bun:test';
import { ALL_BRANCHES, branchOptions } from './branchOptions.js';

const list = [
	{ name: 'main', runCount: 9 },
	{ name: 'feat/x', runCount: 2 },
];

describe('branchOptions', () => {
	test('leads with all branches, then the list in the order given', () => {
		expect(branchOptions(list, null)).toEqual([
			{ value: ALL_BRANCHES, label: 'All branches' },
			{ value: 'main', label: 'main', runCount: 9 },
			{ value: 'feat/x', label: 'feat/x', runCount: 2 },
		]);
	});

	test('keeps a chosen branch that has no runs in this window', () => {
		expect(branchOptions(list, 'old-branch').map((o) => o.value)).toEqual([
			ALL_BRANCHES,
			'main',
			'feat/x',
			'old-branch',
		]);
	});

	test('does not repeat a chosen branch the list already has', () => {
		expect(branchOptions(list, 'main')).toHaveLength(3);
	});

	test('offers only all branches for an empty or missing list', () => {
		expect(branchOptions([], null)).toEqual([
			{ value: ALL_BRANCHES, label: 'All branches' },
		]);
		expect(branchOptions(undefined, null)).toHaveLength(1);
	});
});
