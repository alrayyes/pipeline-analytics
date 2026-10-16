// Options for @codecov/vite-plugin. The upload reuses the CODECOV_TOKEN the
// coverage and test result uploads already use. Without one the plugin does a
// dry run, so a fork, a Dependabot run or a local build still succeeds.
export const bundleAnalysisOptions = (
	env: Record<string, string | undefined>,
) => {
	const uploadToken = env.CODECOV_TOKEN || undefined;
	return {
		enableBundleAnalysis: true,
		bundleName: 'pipeline-analytics',
		uploadToken,
		gitService: 'github' as const,
		// No telemetry to Codecov beyond the upload itself.
		telemetry: false,
		dryRun: uploadToken === undefined,
	};
};
