import { describe, expect, test } from 'bun:test';
import { isPending, outcomeLabel, outcomeTone } from './statusModel.js';

describe('outcomeTone', () => {
	test.each([
		['passed', 'pass'],
		['failed', 'fail'],
		['running', 'running'],
		['queued', 'running'],
		['cancelled', 'neutral'],
		['skipped', 'neutral'],
		['unknown', 'neutral'],
	] as const)('%s is %s', (outcome, tone) => {
		expect(outcomeTone(outcome)).toBe(tone);
	});
});

describe('outcomeLabel', () => {
	test.each([
		['passed', 'Passed'],
		['failed', 'Failed'],
		['running', 'Running'],
		['queued', 'Queued'],
		['cancelled', 'Cancelled'],
		['skipped', 'Skipped'],
	] as const)('%s reads %s', (outcome, label) => {
		expect(outcomeLabel(outcome)).toBe(label);
	});

	test('an unknown outcome shows the forge text it was given, so it is never blank', () => {
		expect(outcomeLabel('unknown', 'action_required')).toBe('action_required');
	});

	test('an unknown outcome with nothing to show says so', () => {
		expect(outcomeLabel('unknown')).toBe('Unknown');
	});

	test('the fallback is only for unknown: a known outcome ignores it', () => {
		expect(outcomeLabel('passed', 'success')).toBe('Passed');
	});
});

describe('isPending', () => {
	test('running and queued are work still to come, so a list should keep refreshing', () => {
		expect(isPending('running')).toBe(true);
		expect(isPending('queued')).toBe(true);
	});

	test.each(['passed', 'failed', 'cancelled', 'skipped', 'unknown'] as const)(
		'%s is not pending',
		(outcome) => {
			expect(isPending(outcome)).toBe(false);
		},
	);
});
