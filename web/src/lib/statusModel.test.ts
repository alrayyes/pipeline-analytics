import { describe, expect, test } from 'bun:test';
import { statusLabel, statusTone } from './statusModel.js';

describe('statusTone', () => {
	test.each([
		['completed', 'success', 'pass'],
		['completed', 'failure', 'fail'],
		['completed', 'timed_out', 'fail'],
		['in_progress', undefined, 'running'],
		['queued', undefined, 'running'],
		['completed', 'cancelled', 'neutral'],
		['completed', 'skipped', 'neutral'],
	] as const)('%s / %s is %s', (status, conclusion, tone) => {
		expect(statusTone(status, conclusion)).toBe(tone);
	});

	test('a conclusion wins over a stale in_progress status', () => {
		expect(statusTone('in_progress', 'failure')).toBe('fail');
	});

	test('an unrecognised state is neutral, not a pass', () => {
		expect(statusTone('weird', 'strange')).toBe('neutral');
	});

	test('a completed run with no conclusion is neutral, not a pass', () => {
		expect(statusTone('completed', undefined)).toBe('neutral');
	});
});

describe('statusLabel', () => {
	test.each([
		['completed', 'success', 'Passed'],
		['completed', 'failure', 'Failed'],
		['completed', 'timed_out', 'Timed out'],
		['completed', 'cancelled', 'Cancelled'],
		['completed', 'skipped', 'Skipped'],
		['in_progress', undefined, 'Running'],
		['queued', undefined, 'Queued'],
	] as const)('%s / %s reads %s', (status, conclusion, label) => {
		expect(statusLabel(status, conclusion)).toBe(label);
	});

	test('an unrecognised state falls back to its raw status text', () => {
		expect(statusLabel('weird', undefined)).toBe('weird');
	});
});
