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
	fetchGitHubRateLimit,
	fetchJobLog,
	fetchPipeline,
	fetchPipelineSteps,
	fetchPipelines,
	fetchRepoUsage,
	fetchRunSteps,
	fetchRuns,
	fetchUnhealthySteps,
	fetchVersion,
	type InsightsWindow,
	quarantineStep,
	type RunStatusFilter,
	unquarantineStep,
} from './dashboardApi.js';
import {
	discoverRepos,
	fetchRepoIdentifiers,
	fetchRepoPage,
	type Forge as RepoForge,
	registerRepo,
	untrackRepo,
} from './repoApi.js';
import {
	actionMessage,
	cancelRun,
	type RunActionKind,
	rerunRun,
} from './runActions.js';
import {
	loadSettings,
	type SettingsUpdate,
	saveSettings,
} from './settingsSync.js';
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

// The spec's enum, checked before a person is asked about a call the server
// would refuse anyway.
function forgeArg(input: Record<string, unknown>): RepoForge {
	const forge = required(input, 'forge');
	if (forge !== 'github' && forge !== 'forgejo') {
		throw new Error('forge must be github or forgejo');
	}

	return forge;
}

async function mustConfirm(
	agent: Agent,
	question: string,
	declined: string,
): Promise<void> {
	if (!(await personConfirms(agent, question))) {
		throw new Error(declined);
	}
}

async function actOnRun(
	kind: RunActionKind,
	input: Record<string, unknown>,
	agent: Agent,
) {
	const runId = required(input, 'runId');
	const verb = kind === 'rerun' ? 're-run' : 'cancel';
	await mustConfirm(
		agent,
		`Let the agent ${verb} run ${runId}?`,
		`The person declined to ${verb} run ${runId}`,
	);

	const result = await (kind === 'rerun' ? rerunRun : cancelRun)(runId);
	if (!result.ok) throw new Error(actionMessage(kind, result));

	return textResult({ runId, action: kind, accepted: true });
}

function textResult(data: unknown) {
	return { content: [{ type: 'text' as const, text: JSON.stringify(data) }] };
}

// The name a person would recognise for a repository id, for the question they
// are asked. The id alone ("r1") means nothing to them. Best effort: a failed,
// slow or empty lookup, or a repository past the first 500, names the id, so the
// confirmation never depends on it. There is no get-by-id endpoint to ask
// instead.
const NAME_LOOKUP_MS = 3000;

async function repoName(repoId: string): Promise<string> {
	try {
		const { repos } = await fetchRepoPage(
			{ limit: 500, offset: 0 },
			(url, init) =>
				fetch(url, { ...init, signal: AbortSignal.timeout(NAME_LOOKUP_MS) }),
		);

		return repos?.find((r) => r.id === repoId)?.identifier ?? repoId;
	} catch {
		return repoId;
	}
}

export function registerWebMCPTools(
	modelContext: ModelContext | undefined = typeof document === 'undefined'
		? undefined
		: document.modelContext,
	// What the page does to show a change a tool made behind its back: the
	// layout reruns its load, which re-seeds the settings stores.
	options: { refreshPage?: () => unknown } = {},
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

	modelContext.registerTool({
		name: 'list_repos',
		description:
			"List the tracked repositories, with each one's forge, ingestion status and masked token. Returns 20 repositories from offset 0 unless limit and offset say otherwise; check hasMore for another page. Use an id from here with untrack_repo.",
		inputSchema: inputSchemas.list_repos,
		execute: async (input) =>
			textResult(
				await fetchRepoPage({
					limit: typeof input.limit === 'number' ? input.limit : 20,
					offset: typeof input.offset === 'number' ? input.offset : 0,
					forge: input.forge as RepoForge | undefined,
				}),
			),
	});

	modelContext.registerTool({
		name: 'list_repo_identifiers',
		description:
			'List every tracked repository of one forge as owner/name, unpaginated, to see whether one is already tracked. For Forgejo, narrow it to one instance.',
		inputSchema: inputSchemas.list_repo_identifiers,
		execute: async (input) =>
			textResult({
				identifiers: await fetchRepoIdentifiers({
					forge: forgeArg(input),
					forgejoInstanceUrl:
						typeof input.forgejoInstanceUrl === 'string'
							? input.forgejoInstanceUrl
							: undefined,
				}),
			}),
	});

	modelContext.registerTool({
		name: 'discover_repos',
		description:
			'List the repositories a token can reach on a forge, as owner/name to pass to register_repo. Omit the token to use the one saved for the forge and instance; a token you pass is used for this one lookup and never stored. Nothing is registered.',
		inputSchema: inputSchemas.discover_repos,
		execute: async (input) =>
			textResult({
				identifiers: await discoverRepos({
					forge: forgeArg(input),
					forgejoInstanceUrl:
						typeof input.forgejoInstanceUrl === 'string'
							? input.forgejoInstanceUrl
							: undefined,
					token: typeof input.token === 'string' ? input.token : undefined,
				}),
			}),
	});

	modelContext.registerTool({
		name: 'register_repo',
		description:
			'Start tracking a repository: store its token encrypted, create a webhook on the forge and begin ingesting its runs. Asks the person to confirm first and does nothing if they decline. Omit the token to use the one saved for the forge and instance. Refused for an archived repository, a fork or a mirror, or one already tracked. Takes an identifier from discover_repos.',
		inputSchema: inputSchemas.register_repo,
		annotations: { destructiveHint: true },
		execute: async (input, agent) => {
			const forge = forgeArg(input);
			const identifier = required(input, 'identifier');
			await mustConfirm(
				agent,
				`Let the agent start tracking ${identifier} on ${forge}? This creates a webhook on the repository.`,
				`The person declined to track ${identifier}`,
			);

			return textResult(
				await registerRepo({
					forge,
					identifier,
					forgejoInstanceUrl:
						typeof input.forgejoInstanceUrl === 'string'
							? input.forgejoInstanceUrl
							: undefined,
					token: typeof input.token === 'string' ? input.token : undefined,
				}),
			);
		},
	});

	modelContext.registerTool({
		name: 'untrack_repo',
		description:
			'Stop tracking a repository and delete its stored runs, jobs and steps. The webhook on the forge stays. Asks the person to confirm first and does nothing if they decline. Takes a repository id from list_repos.',
		inputSchema: inputSchemas.untrack_repo,
		annotations: { destructiveHint: true },
		execute: async (input, agent) => {
			const repoId = required(input, 'repoId');
			const name = await repoName(repoId);
			await mustConfirm(
				agent,
				`Let the agent stop tracking repository ${name}? Its stored runs, jobs and steps are deleted.`,
				`The person declined to untrack repository ${name}`,
			);
			await untrackRepo(repoId);

			return textResult({ repoId, untracked: true });
		},
	});

	modelContext.registerTool({
		name: 'get_version',
		description:
			'Get the version of the running service, the one the page footer shows: a tagged release such as v1.2.3, or dev for an untagged local build.',
		inputSchema: inputSchemas.get_version,
		execute: async () => textResult(await fetchVersion()),
	});

	modelContext.registerTool({
		name: 'get_github_rate_limit',
		description:
			'Get, for each GitHub token this service holds, the repositories it is used for and its most recently observed rate-limit status. The status comes from real API responses, so it is absent for a token not yet used. Tokens are masked to their last four characters.',
		inputSchema: inputSchemas.get_github_rate_limit,
		execute: async () => textResult({ tokens: await fetchGitHubRateLimit() }),
	});

	modelContext.registerTool({
		name: 'get_settings',
		description:
			"Get the signed-in account's dashboard settings: the theme, the forge filter, the Pipelines page's filters and sort, and the telemetry window, each with the default it falls back to.",
		inputSchema: inputSchemas.get_settings,
		execute: async () => textResult(await loadSettings()),
	});

	modelContext.registerTool({
		name: 'update_settings',
		description:
			"Change the signed-in account's dashboard settings, the same ones the page's own controls save. Give only the settings to change: a value sets it, null clears it back to its default, and one left out is untouched. The server rejects the whole update if any value is invalid. Returns the settings after the change, and the open page picks them up.",
		inputSchema: inputSchemas.update_settings,
		execute: async (input) => {
			const saved = await saveSettings(input as SettingsUpdate);
			try {
				await options.refreshPage?.();
			} catch {
				// The change is saved; the page shows it on its next navigation.
			}

			return textResult(saved);
		},
	});
}
