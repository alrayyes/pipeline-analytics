import { describe, expect, test } from 'bun:test';
import { flakyRunsHref } from './flakyRuns.js';

describe('flakyRunsHref', () => {
	test('links a step to its pipeline-scoped flaky-runs list', () => {
		expect(flakyRunsHref('pipeline-1', 'run tests')).toBe(
			'/pipelines/pipeline-1/flaky-runs?step=run%20tests',
		);
	});

	test('encodes characters the step name may contain', () => {
		expect(flakyRunsHref('pipeline-1', 'lint & test / build')).toBe(
			'/pipelines/pipeline-1/flaky-runs?step=lint%20%26%20test%20%2F%20build',
		);
	});
});
