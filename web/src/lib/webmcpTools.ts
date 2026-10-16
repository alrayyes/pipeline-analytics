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
	fetchFlakyRuns,
	fetchFlakySteps,
	fetchJobLog,
	fetchPipeline,
	fetchPipelineSteps,
	fetchPipelines,
	fetchRepoUsage,
	fetchRunSteps,
	fetchRuns,
	fetchUnhealthySteps,
	type InsightsWindow,
	quarantineStep,
	type RunStatusFilter,
	unquarantineStep,
} from './dashboardApi.js';
import {
	actionMessage,
	cancelRun,
	type RunActionKind,
	rerunRun,
} from './runActions.js';
import { inputSchemas } from './webmcpSchemas.generated.js';

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

// A missing required argument is an error the agent can read, not the string
// "undefined" sent to the server, which answers an empty list or a 404.
function required(input: Record<string, unknown>, key: string): string {
	const value = input[key];
	if (typeof value !== 'string' || value === '') {
		throw new Error(`${key} is required`);
	}

	return value;
}

type Agent = Parameters<
	NonNullable<Parameters<ModelContext['registerTool']>[0]['execute']>
>[1];

// A write the person has not just said yes to does not happen. The question is
// put through the agent's requestUserInteraction when it has one, and a browser
// confirm() otherwise; a page with neither answers no, never yes. This is the
// Runs page's "Cancel run?" dialog for a caller that has no click to give.
async function personConfirms(
	agent: Agent,
	question: string,
): Promise<boolean> {
	const ask = () =>
		typeof globalThis.confirm === 'function' && globalThis.confirm(question);

	// Only a plain true is a yes: a promise, a string or an object is not.
	const answer = await (agent.requestUserInteraction
		? agent.requestUserInteraction(ask)
		: ask());

	return answer === true;
}

async function actOnRun(
	kind: RunActionKind,
	input: Record<string, unknown>,
	agent: Agent,
) {
	const runId = required(input, 'runId');
	const verb = kind === 'rerun' ? 're-run' : 'cancel';
	if (!(await personConfirms(agent, `Let the agent ${verb} run ${runId}?`))) {
		throw new Error(`The person declined to ${verb} run ${runId}`);
	}

	const result = await (kind === 'rerun' ? rerunRun : cancelRun)(runId);
	if (!result.ok) throw new Error(actionMessage(kind, result));

	return textResult({ runId, action: kind, accepted: true });
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
			"List tracked pipelines, grouped by repository, with each one's current health status. Returns 20 pipelines from offset 0 unless limit and offset say otherwise; check hasMore for another page. Use a pipelineId from here with get_pipeline.",
		inputSchema: inputSchemas.list_pipelines,
		execute: async (input) => {
			const data = await fetchPipelines({
				limit: typeof input.limit === 'number' ? input.limit : 20,
				offset: typeof input.offset === 'number' ? input.offset : 0,
				forge: input.forge as 'github' | 'forgejo' | undefined,
				repoId: input.repoId as string | undefined,
				health: input.health as 'healthy' | 'unhealthy' | undefined,
				sort: input.sort as 'name' | 'lastRun' | undefined,
				window: input.window as string | undefined,
			});

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_pipeline',
		description:
			"Get one pipeline's health status and its duration and failure-rate trends. Takes a pipelineId from list_pipelines.",
		inputSchema: inputSchemas.get_pipeline,
		execute: async (input) => {
			const data = await fetchPipeline(input.pipelineId as string, undefined, {
				window: input.window as string | undefined,
			});

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_repo_usage',
		description:
			"Get one repository's runner-minutes usage, broken down by workflow. Takes a repoId from list_pipelines.",
		inputSchema: inputSchemas.get_repo_usage,
		execute: async (input) => {
			const data = await fetchRepoUsage(input.repoId as string, undefined, {
				window: input.window as string | undefined,
			});

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'list_pipeline_steps',
		description:
			"List one pipeline's steps with their outcome counts, to find where it fails. Takes a pipelineId from list_pipelines.",
		inputSchema: inputSchemas.list_pipeline_steps,
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
			"Get the last lines of one job's log from a run, to read why a step failed. Takes the runId and the failed step's jobId from get_run_steps.",
		inputSchema: inputSchemas.get_job_log,
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
			'Get what failed over a window (24h, 7d or 30d; 7d when omitted): the failure rate, the steps that fail most, the categories and the pipelines. Narrow it by forge, repoId or branch.',
		inputSchema: inputSchemas.get_failure_insights,
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
			'List recent runs, newest first, filtered by status, repository, forge or branch. Returns 20 runs from offset 0 unless limit and offset say otherwise; check hasMore for another page. Open one with get_run_steps.',
		inputSchema: inputSchemas.list_runs,
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
			'List the steps that flake over a window, flakiest first, with their quarantine state. Narrow it by forge, repoId or branch.',
		inputSchema: inputSchemas.list_flaky_steps,
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
			'List the branches that ran over a window, with their run counts, to find a branch value for the other tools.',
		inputSchema: inputSchemas.list_branches,
		execute: async (input) =>
			textResult(
				await fetchBranches({
					window: input.window as InsightsWindow | undefined,
					forge: input.forge as Forge | undefined,
					repoId: input.repoId as string | undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'get_run_steps',
		description:
			"Get one run's steps with their outcomes and, for a failed step, its job id for get_job_log. Takes a runId from list_runs.",
		inputSchema: inputSchemas.get_run_steps,
		execute: async (input) =>
			textResult(await fetchRunSteps(required(input, 'runId'))),
	});

	modelContext.registerTool({
		name: 'list_pipeline_flaky_runs',
		description:
			'List the runs in which one step of a pipeline flaked, to open them with get_run_steps. Takes a pipelineId and a step name from list_flaky_steps.',
		inputSchema: inputSchemas.list_pipeline_flaky_runs,
		execute: async (input) =>
			textResult(
				await fetchFlakyRuns(required(input, 'pipelineId'), {
					step: required(input, 'step'),
					window: typeof input.window === 'string' ? input.window : undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'list_unhealthy_steps',
		description:
			'List the pipelines that have a flaky or failing step over a window, with those steps. Pages with limit and offset.',
		inputSchema: inputSchemas.list_unhealthy_steps,
		execute: async (input) =>
			textResult(
				await fetchUnhealthySteps({
					window: typeof input.window === 'string' ? input.window : undefined,
					limit: typeof input.limit === 'number' ? input.limit : undefined,
					offset: typeof input.offset === 'number' ? input.offset : undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'rerun_run',
		description:
			'Ask the forge to re-run a finished run, from the run list. Asks the person to confirm first and does nothing if they decline. The forge acts on its own time, so accepted means it took the request, not that the run has started.',
		inputSchema: inputSchemas.rerun_run,
		annotations: { destructiveHint: true },
		execute: (input, agent) => actOnRun('rerun', input, agent),
	});

	modelContext.registerTool({
		name: 'cancel_run',
		description:
			'Ask the forge to cancel a run that has not finished, from the run list. Asks the person to confirm first and does nothing if they decline. The forge acts on its own time, so accepted means it took the request, not that the run has stopped.',
		inputSchema: inputSchemas.cancel_run,
		annotations: { destructiveHint: true },
		execute: (input, agent) => actOnRun('cancel', input, agent),
	});

	modelContext.registerTool({
		name: 'quarantine_step',
		description:
			'Mark a flaky step as known for 30 days, with an optional note, so it stops making its pipeline unhealthy. Marking a step already quarantined renews it. Recorded in this service only; no forge is contacted. Takes a pipelineId and a step name from list_flaky_steps. The signed-in session is needed, as for the Flaky page button.',
		inputSchema: inputSchemas.quarantine_step,
		execute: async (input) =>
			textResult(
				await quarantineStep(
					required(input, 'pipelineId'),
					required(input, 'step'),
					typeof input.note === 'string' ? input.note : undefined,
				),
			),
	});

	modelContext.registerTool({
		name: 'unquarantine_step',
		description:
			"Clear a step's quarantine so it raises the flaky_step health signal again. A step with no mark is fine. Takes a pipelineId and a step name from list_flaky_steps.",
		inputSchema: inputSchemas.unquarantine_step,
		execute: async (input) =>
			textResult(
				await unquarantineStep(
					required(input, 'pipelineId'),
					required(input, 'step'),
				),
			),
	});
}
