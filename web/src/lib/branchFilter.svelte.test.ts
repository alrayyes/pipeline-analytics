import { beforeEach, describe, expect, test } from 'bun:test';
import { clearBranch, getBranch, setBranch } from './branchFilter.svelte.js';

beforeEach(() => clearBranch());

describe('the chosen branch', () => {
	test('starts as no selection, which means every branch', () => {
		expect(getBranch()).toBeNull();
	});

	test('a set branch is read back, and clearing returns to no selection', () => {
		setBranch('main');
		expect(getBranch()).toBe('main');

		clearBranch();
		expect(getBranch()).toBeNull();
	});

	test('an empty name is no selection', () => {
		setBranch('main');
		setBranch('');

		expect(getBranch()).toBeNull();
	});
});
