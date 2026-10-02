import { describe, expect, test } from 'bun:test';
import { stageSegments, stageSummary } from './stageProgress.js';

const step = (name: string, status: string, conclusion?: string) => ({
	name,
	status,
	conclusion,
});

describe('stageSegments', () => {
	test('maps each step to a segment with its tone and a text label', () => {
		const segments = stageSegments([
			step('checkout', 'completed', 'success'),
			step('canary', 'completed', 'failure'),
			step('promote', 'queued'),
		]);

		expect(segments).toEqual([
			{ name: 'checkout', tone: 'pass', label: 'Passed' },
			{ name: 'canary', tone: 'fail', label: 'Failed' },
			{ name: 'promote', tone: 'running', label: 'Queued' },
		]);
	});

	test('no steps is no segments', () => {
		expect(stageSegments([])).toEqual([]);
	});
});

describe('stageSummary', () => {
	test('names the first failing stage and its position', () => {
		expect(
			stageSummary([
				step('checkout', 'completed', 'success'),
				step('build', 'completed', 'success'),
				step('canary', 'completed', 'failure'),
				step('promote', 'queued'),
			]),
		).toBe('Stage 3/4: canary');
	});

	test('a failure outranks a later running stage', () => {
		expect(
			stageSummary([
				step('a', 'completed', 'failure'),
				step('b', 'in_progress'),
			]),
		).toBe('Stage 1/2: a');
	});

	test('names the first running stage when nothing has failed', () => {
		expect(
			stageSummary([
				step('install', 'completed', 'success'),
				step('e2e', 'in_progress'),
				step('upload', 'queued'),
			]),
		).toBe('Stage 2/3: e2e');
	});

	test('all passed reads as a count', () => {
		expect(
			stageSummary([
				step('init', 'completed', 'success'),
				step('plan', 'completed', 'success'),
			]),
		).toBe('2/2 passed');
	});

	test('a skipped step is not counted as passed', () => {
		expect(
			stageSummary([
				step('init', 'completed', 'success'),
				step('lint', 'completed', 'skipped'),
			]),
		).toBe('1/2 passed');
	});

	test('no steps says so rather than 0/0', () => {
		expect(stageSummary([])).toBe('No steps recorded');
	});
});
