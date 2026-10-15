import { describe, expect, mock, test } from 'bun:test';

// Parity between openapi.yaml and the WebMCP tools (rules/mcp.md): every
// operation is a tool, carries an x-agent-excluded reason in the spec, or is
// on the temporary pending list below. The MCP server's twin is
// internal/httpserver/mcp_spec_parity_test.go, which reads the same
// extension. Per-tool behaviour is webmcpTools.test.ts and the API's own tests.

mock.module('./dashboardApi.js', () => ({
	fetchPipelines: () => Promise.resolve({}),
	fetchPipeline: () => Promise.resolve({}),
	fetchRepoUsage: () => Promise.resolve({}),
	fetchPipelineSteps: () => Promise.resolve([]),
	fetchJobLog: () => Promise.resolve({}),
	fetchFailureInsights: () => Promise.resolve({}),
	fetchRuns: () => Promise.resolve({}),
	fetchFlakySteps: () => Promise.resolve({}),
	fetchBranches: () => Promise.resolve({}),
}));

const { registerWebMCPTools } = await import('./webmcpTools.js');

// Each registered WebMCP tool and the operations it covers. A tool may fold
// several operations; it may not leave one out.
const toolOperations: Record<string, string[]> = {
	list_pipelines: ['listPipelines'],
	get_pipeline: ['getPipeline'],
	get_repo_usage: ['getRepoUsage'],
	list_pipeline_steps: ['getPipelineSteps'],
	get_job_log: ['getJobLog'],
	get_failure_insights: ['getFailureInsights'],
	list_runs: ['listRuns'],
	list_flaky_steps: ['listFlakySteps'],
	list_branches: ['listBranches'],
};

// Operations whose WebMCP tool is still to come under #615. It exists so the
// parity test can land first; each tool pull request takes its operations off
// and the last one deletes the list. "Nobody asked yet" is not an exclusion.
const pendingOperations = new Set<string>([
	'getVersion',
	'listRepos',
	'registerRepo',
	'listRepoIdentifiers',
	'discoverRepos',
	'untrackRepo',
	'quarantineStep',
	'unquarantineStep',
	'listFlakyRuns',
	'getRunSteps',
	'rerunRun',
	'cancelRun',
	'getGitHubRateLimitInsights',
	'listUnhealthySteps',
	'getSettings',
	'updateSettings',
]);

const methods = ['get', 'post', 'put', 'patch', 'delete'] as const;

type Operation = { operationId: string; 'x-agent-excluded'?: unknown };

async function specOperations(): Promise<Operation[]> {
	const text = await Bun.file(
		new URL('../../../openapi/openapi.yaml', import.meta.url),
	).text();
	const spec = Bun.YAML.parse(text) as {
		paths: Record<string, Partial<Record<(typeof methods)[number], Operation>>>;
	};

	return Object.values(spec.paths).flatMap((item) =>
		methods.flatMap((method) => (item[method] ? [item[method]] : [])),
	);
}

function registeredToolNames(): string[] {
	const names: string[] = [];
	registerWebMCPTools({
		registerTool: (tool: { name: string }) => {
			names.push(tool.name);
		},
	} as unknown as NonNullable<Document['modelContext']>);

	return names;
}

// Stryker copies only web/ into its sandbox, so the spec one level above is not
// there and every test here would fail its dry run. It sets this variable in
// its workers; the frontend test job, which has the whole checkout, still runs
// these. There is no source here for a mutant to live in anyway.
const inStryker = process.env.STRYKER_MUTATOR_WORKER !== undefined;

describe.skipIf(inStryker)('WebMCP parity with openapi.yaml', () => {
	test('every operation has an operationId', async () => {
		const unnamed = (await specOperations()).filter(
			(op) => typeof op.operationId !== 'string' || op.operationId === '',
		);

		expect(unnamed).toEqual([]);
	});

	test('every operation has a tool, an exclusion reason or a pending entry', async () => {
		const covered = new Set(Object.values(toolOperations).flat());
		const offenders = (await specOperations())
			.filter((op) => !op['x-agent-excluded'])
			.map((op) => op.operationId)
			.filter((id) => !covered.has(id) && !pendingOperations.has(id));

		expect(offenders).toEqual([]);
	});

	test('an exclusion states its reason', async () => {
		const blank = (await specOperations())
			.filter((op) => 'x-agent-excluded' in op)
			.filter((op) => {
				const reason = op['x-agent-excluded'];

				return typeof reason !== 'string' || reason.trim() === '';
			})
			.map((op) => op.operationId);

		expect(blank).toEqual([]);
	});

	test('every registered tool is mapped to operations the spec declares', async () => {
		const declared = new Set(
			(await specOperations()).map((op) => op.operationId),
		);
		const unmapped = registeredToolNames().filter(
			(name) => !toolOperations[name],
		);
		const unknown = Object.values(toolOperations)
			.flat()
			.filter((id) => !declared.has(id));

		expect({ unmapped, unknown }).toEqual({ unmapped: [], unknown: [] });
	});

	test('every mapped tool is registered', () => {
		const registered = new Set(registeredToolNames());
		const missing = Object.keys(toolOperations).filter(
			(name) => !registered.has(name),
		);

		expect(missing).toEqual([]);
	});

	test('the pending list holds only operations still without a tool', async () => {
		const covered = new Set(Object.values(toolOperations).flat());
		const excluded = new Set(
			(await specOperations())
				.filter((op) => op['x-agent-excluded'])
				.map((op) => op.operationId),
		);
		const declared = new Set(
			(await specOperations()).map((op) => op.operationId),
		);
		const stale = [...pendingOperations].filter(
			(id) => covered.has(id) || excluded.has(id) || !declared.has(id),
		);

		expect(stale).toEqual([]);
	});
});
