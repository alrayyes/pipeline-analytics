import { expect, test } from 'bun:test';
import { bundleAnalysisOptions } from './bundleAnalysis';

// The bundle analysis upload reuses CODECOV_TOKEN (#421). A build with no
// token (a fork, a Dependabot run, a local build) must still succeed, so the
// plugin is told not to upload rather than left to fail on a missing token.
test('uploads with the token from the environment', () => {
	expect(bundleAnalysisOptions({ CODECOV_TOKEN: 'abc' })).toEqual({
		enableBundleAnalysis: true,
		bundleName: 'pipeline-analytics',
		uploadToken: 'abc',
		gitService: 'github',
		telemetry: false,
		dryRun: false,
	});
});

test('does a dry run without a token', () => {
	expect(bundleAnalysisOptions({}).dryRun).toBe(true);
	expect(bundleAnalysisOptions({ CODECOV_TOKEN: '' }).dryRun).toBe(true);
});
