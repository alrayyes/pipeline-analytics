// Registers the dashboard's read views as WebMCP tools
// (https://github.com/webmachinelearning/webmcp) for an in-browser agent
// sharing the page with its user (#296). Mirrors serviceWorker.ts's shape:
// a feature check, then a fire-and-forget registration that's a no-op on
// any browser that doesn't support the API.
//
// Every tool's execute() calls the exact same function the corresponding
// page uses to fetch its data, rather than a second implementation of the
// same request -- see design.md's "second path to the same API" decision.

import {
	fetchBranches,
	fetchFailureInsights,
	fetchFlakySteps,
	fetchJobLog,
	fetchPipeline,
	fetchPipelineSteps,
	fetchPipelines,
	fetchRepoUsage,
	fetchRuns,
	type InsightsWindow,
	type RunStatusFilter,
} from './dashboardApi.js';

type ModelContext = NonNullable<Document['modelContext']>;

type Forge = 'github' | 'forgejo';

// The filters the telemetry reads share: the server owns every default, so a
// field the agent leaves out stays out of the request.
function scope(input: Record<string, unknown>) {
	return {
		forge: input.forge as Forge | undefined,
		repoId: input.repoId as string | undefined,
		branch: input.branch as string | undefined,
	};
}

function textResult(data: unknown) {
	return { content: [{ type: 'text' as const, text: JSON.stringify(data) }] };
}

export function registerWebMCPTools(
	modelContext: ModelContext | undefined = typeof document === 'undefined'
		? undefined
		: document.modelContext,
): void {
	if (!modelContext) return;

	modelContext.registerTool({
		name: 'list_pipelines',
		description:
			'List tracked pipelines, grouped by repository, with their current health status.',
		execute: async (input) => {
			const data = await fetchPipelines({
				limit: typeof input.limit === 'number' ? input.limit : 20,
				offset: typeof input.offset === 'number' ? input.offset : 0,
				forge: input.forge as 'github' | 'forgejo' | undefined,
				repoId: input.repoId as string | undefined,
			});

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_pipeline',
		description:
			"Get a pipeline's health status and its duration and failure-rate trends.",
		execute: async (input) => {
			const data = await fetchPipeline(input.pipelineId as string);

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_repo_usage',
		description:
			"Get a repository's runner-minutes usage, broken down by workflow.",
		execute: async (input) => {
			const data = await fetchRepoUsage(input.repoId as string);

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'list_pipeline_steps',
		description:
			"List a pipeline's steps with their outcome counts, to find where it fails.",
		execute: async (input) =>
			textResult(
				await fetchPipelineSteps(input.pipelineId as string, {
					window: input.window as InsightsWindow | undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'get_job_log',
		description:
			"Get the last lines of one job's log from a run, to read why a step failed.",
		execute: async (input) =>
			textResult(
				await fetchJobLog(input.runId as string, input.jobId as string, {
					lines: typeof input.lines === 'number' ? input.lines : undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'get_failure_insights',
		description:
			'Get what failed over a window: failure rate, the steps that fail most, the categories and the pipelines.',
		execute: async (input) =>
			textResult(
				await fetchFailureInsights({
					window: input.window as InsightsWindow | undefined,
					forge: input.forge as Forge | undefined,
					repoId: input.repoId as string | undefined,
					branch: input.branch as string | undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'list_runs',
		description:
			'List recent runs, newest first, filtered by status, repository, forge or branch.',
		execute: async (input) =>
			textResult(
				await fetchRuns({
					limit: typeof input.limit === 'number' ? input.limit : 20,
					offset: typeof input.offset === 'number' ? input.offset : 0,
					status: input.status as RunStatusFilter | undefined,
					...scope(input),
				}),
			),
	});

	modelContext.registerTool({
		name: 'list_flaky_steps',
		description:
			'List the steps that flake over a window, flakiest first, with their quarantine state.',
		execute: async (input) =>
			textResult(
				await fetchFlakySteps({
					window: input.window as InsightsWindow | undefined,
					limit: typeof input.limit === 'number' ? input.limit : undefined,
					offset: typeof input.offset === 'number' ? input.offset : undefined,
					...scope(input),
				}),
			),
	});

	modelContext.registerTool({
		name: 'list_branches',
		description:
			'List the branches that ran over a window, with their run counts.',
		execute: async (input) =>
			textResult(
				await fetchBranches({
					window: input.window as InsightsWindow | undefined,
					forge: input.forge as Forge | undefined,
					repoId: input.repoId as string | undefined,
				}),
			),
	});
}
