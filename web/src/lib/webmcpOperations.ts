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
	rerun_run: 'rerunRun',
	cancel_run: 'cancelRun',
	quarantine_step: 'quarantineStep',
	unquarantine_step: 'unquarantineStep',
	list_repos: 'listRepos',
	list_repo_identifiers: 'listRepoIdentifiers',
	register_repo: 'registerRepo',
	discover_repos: 'discoverRepos',
	untrack_repo: 'untrackRepo',
	get_settings: 'getSettings',
	update_settings: 'updateSettings',
	get_version: 'getVersion',
	get_github_rate_limit: 'getGitHubRateLimitInsights',
};

// Where a tool pages by default and the API doesn't. The spec says an omitted
// `limit` returns everything, which is right for the endpoint and wrong for an
// agent, so the tool's schema states its own default (#649).
// The spec's wording is for raw HTTP callers, who encode a `/` themselves; the
// tool encodes it, so an agent that did too would address a different step.
const STEP_AS_LISTED =
	"The step's name as list_flaky_steps shows it, for example `lint/css`. Do not encode it.";

// Same reason for a repo id: the tool encodes it, so the agent passes it as
// list_repos reports it.
const REPO_AS_LISTED =
	"The repository's id as list_repos reports it. Do not encode it.";

const INSTANCE =
	"For Forgejo, the instance's URL, for example `https://git.example.com`. Not used for GitHub.";

const CLEARS = ' null clears it back to its default.';

export const toolParameterOverrides: Record<
	string,
	Record<string, { default?: number; description: string }>
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
	list_repos: {
		limit: {
			default: 20,
			description:
				'Max repositories to return. Defaults to 20 here; the API itself returns every repository when it is omitted.',
		},
		offset: {
			default: 0,
			description: 'Repositories to skip before the returned page.',
		},
	},
	register_repo: {
		identifier: {
			description:
				'The repository as owner/name, for example `alrayyes/pipeline-analytics`, as discover_repos lists it. At most 200 characters.',
		},
		forgejoInstanceUrl: { description: INSTANCE },
	},
	discover_repos: { forgejoInstanceUrl: { description: INSTANCE } },
	list_repo_identifiers: { forgejoInstanceUrl: { description: INSTANCE } },
	update_settings: {
		theme: {
			description: `The colour theme: light, dark, or system to follow the device.${CLEARS}`,
		},
		forgeFilter: {
			description: `Which forge the dashboard shows: all, github or forgejo.${CLEARS}`,
		},
		pipelinesHealthFilter: {
			description: `The Pipelines page's health filter: all, healthy or unhealthy.${CLEARS}`,
		},
		pipelinesRepoSelector: {
			description: `The repository the Pipelines page is limited to, as an id from list_repos, or all for every repository.${CLEARS}`,
		},
		pipelinesSortOrder: {
			description: `How the Pipelines page sorts: name, or lastRun for the most recently run first.${CLEARS}`,
		},
		telemetryWindow: {
			description: `The window the telemetry pages cover: 24h, 7d or 30d.${CLEARS}`,
		},
	},
	untrack_repo: { repoId: { description: REPO_AS_LISTED } },
	quarantine_step: { step: { description: STEP_AS_LISTED } },
	unquarantine_step: { step: { description: STEP_AS_LISTED } },
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
