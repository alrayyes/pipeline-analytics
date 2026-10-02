import { describe, expect, test } from 'bun:test';
import type { Outcome } from './dashboardApi.js';
import { stageSegments, stageSummary } from './stageProgress.js';

// The forge's raw status and conclusion ride along but are never read: the
// outcome is all the mapping looks at.
const step = (name: string, outcome: Outcome) => ({
	name,
	status: 'ignored',
	conclusion: 'ignored',
	outcome,
});

describe('stageSegments', () => {
	test('maps each step to a segment with its tone and a text label', () => {
		const segments = stageSegments([
			step('checkout', 'passed'),
			step('canary', 'failed'),
			step('promote', 'queued'),
		]);

		expect(segments).toEqual([
			{ name: 'checkout', tone: 'pass', label: 'Passed' },
			{ name: 'canary', tone: 'fail', label: 'Failed' },
			{ name: 'promote', tone: 'running', label: 'Queued' },
		]);
	});

	test('an unknown step shows its forge text as the label', () => {
		const [segment] = stageSegments([
			{
				name: 'gate',
				status: 'completed',
				conclusion: 'action_required',
				outcome: 'unknown',
			},
		]);

		expect(segment.label).toBe('action_required');
	});

	test('no steps is no segments', () => {
		expect(stageSegments([])).toEqual([]);
	});
});

describe('stageSummary', () => {
	test('names the first failing stage and its position', () => {
		expect(
			stageSummary([
				step('checkout', 'passed'),
				step('build', 'passed'),
				step('canary', 'failed'),
				step('promote', 'queued'),
			]),
		).toBe('Stage 3/4: canary');
	});

	test('a failure outranks a later running stage', () => {
		expect(stageSummary([step('a', 'failed'), step('b', 'running')])).toBe(
			'Stage 1/2: a',
		);
	});

	test('names the first running stage when nothing has failed', () => {
		expect(
			stageSummary([
				step('install', 'passed'),
				step('e2e', 'running'),
				step('upload', 'queued'),
			]),
		).toBe('Stage 2/3: e2e');
	});

	test('all passed reads as a count', () => {
		expect(stageSummary([step('init', 'passed'), step('plan', 'passed')])).toBe(
			'2/2 passed',
		);
	});

	test('a skipped step is not counted as passed', () => {
		expect(
			stageSummary([step('init', 'passed'), step('lint', 'skipped')]),
		).toBe('1/2 passed');
	});

	test('no steps says so rather than 0/0', () => {
		expect(stageSummary([])).toBe('No steps recorded');
	});
});
