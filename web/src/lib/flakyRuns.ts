// A flaky step's aggregate forgeUrl is one arbitrarily-picked occurrence
// (#216) -- often a run that's since passed -- so a flaky step routes to
// the runs it actually failed on instead of linking straight out.
export function flakyRunsHref(pipelineId: string, stepName: string): string {
	return `/pipelines/${pipelineId}/flaky-runs?step=${encodeURIComponent(stepName)}`;
}
