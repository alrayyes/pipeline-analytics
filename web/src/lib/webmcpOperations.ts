// The WebMCP tools and the OpenAPI operation each one wraps. The input schemas
// are generated from these operations (scripts/generate-webmcp-schemas.ts), and
// webmcpParity.test.ts checks every operation in the spec against them
// (rules/mcp.md). One operation per tool for now; a tool that folds several
// would name its schema's operation here and the rest in the parity test.

export const toolOperations: Record<string, string> = {
	list_pipelines: 'listPipelines',
	get_pipeline: 'getPipeline',
	get_repo_usage: 'getRepoUsage',
	list_pipeline_steps: 'getPipelineSteps',
	get_job_log: 'getJobLog',
	get_failure_insights: 'getFailureInsights',
	list_runs: 'listRuns',
	list_flaky_steps: 'listFlakySteps',
	list_branches: 'listBranches',
	get_run_steps: 'getRunSteps',
	list_pipeline_flaky_runs: 'listFlakyRuns',
	list_unhealthy_steps: 'listUnhealthySteps',
};

// Where a tool pages by default and the API doesn't. The spec says an omitted
// `limit` returns everything, which is right for the endpoint and wrong for an
// agent, so the tool's schema states its own default (#649).
export const toolParameterOverrides: Record<
	string,
	Record<string, { default: number; description: string }>
> = {
	list_pipelines: {
		limit: {
			default: 20,
			description:
				'Max pipelines to return. Defaults to 20 here; the API itself returns every pipeline when it is omitted.',
		},
		offset: {
			default: 0,
			description: 'Pipelines to skip before the returned page.',
		},
	},
	list_runs: {
		limit: {
			default: 20,
			description:
				'Max runs to return. Defaults to 20 here; the API itself returns every run when it is omitted.',
		},
		offset: {
			default: 0,
			description: 'Runs to skip before the returned page.',
		},
	},
};
