import { describe, expect, test } from 'bun:test';
import type { FlakyStep, Outcome } from './dashboardApi.js';
import { flakeCells, flakeSummary } from './flakeMatrix.js';

const step = (
	recentOutcomes: Outcome[],
	overrides: Partial<FlakyStep> = {},
): FlakyStep => ({
	pipelineId: 'p1',
	pipelineName: 'CI',
	repoId: 'r1',
	name: 'test',
	flakeRate: 0.25,
	runCount: recentOutcomes.length,
	recentOutcomes,
	...overrides,
});

describe('flakeCells', () => {
	test('gives each result a tone and a text label, oldest first', () => {
		expect(flakeCells(['passed', 'failed', 'cancelled'])).toEqual([
			{ tone: 'pass', label: 'Passed' },
			{ tone: 'fail', label: 'Failed' },
			{ tone: 'neutral', label: 'Cancelled' },
		]);
	});

	test('an outcome the server could not classify reads Unknown, never a pass', () => {
		expect(flakeCells(['unknown'])).toEqual([
			{ tone: 'neutral', label: 'Unknown' },
		]);
	});

	test('an empty history is an empty matrix', () => {
		expect(flakeCells([])).toEqual([]);
	});
});

describe('flakeSummary', () => {
	test('says how many of the shown runs failed, and the rate over the window', () => {
		const summary = flakeSummary(
			step(['passed', 'failed', 'failed', 'passed'], {
				flakeRate: 0.5,
				runCount: 4,
			}),
		);

		expect(summary).toBe(
			'2 of the last 4 runs failed. 50% failure rate over 4 runs.',
		);
	});

	test('names the full run count when the matrix shows only the most recent', () => {
		const shown = Array.from(
			{ length: 40 },
			(_, i): Outcome => (i === 39 ? 'failed' : 'passed'),
		);

		expect(flakeSummary(step(shown, { flakeRate: 0.2, runCount: 150 }))).toBe(
			'1 of the last 40 runs failed. 20% failure rate over 150 runs.',
		);
	});

	test('a single run reads in the singular', () => {
		expect(flakeSummary(step(['failed'], { flakeRate: 1, runCount: 1 }))).toBe(
			'1 of the last 1 run failed. 100% failure rate over 1 run.',
		);
	});
});
