import { describe, expect, mock, test } from 'bun:test';

// Parity between openapi.yaml and the WebMCP tools (rules/mcp.md): every
// operation is a tool or carries an x-agent-excluded reason in the spec, with
// no third list to park one on. The MCP server's twin is
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
	fetchRunSteps: () => Promise.resolve({}),
	fetchFlakyRuns: () => Promise.resolve([]),
	fetchUnhealthySteps: () => Promise.resolve({}),
	quarantineStep: () => Promise.resolve({}),
	unquarantineStep: () => Promise.resolve({}),
	fetchVersion: () => Promise.resolve({}),
	fetchGitHubRateLimit: () => Promise.resolve([]),
}));

const { registerWebMCPTools } = await import('./webmcpTools.js');
const { toolOperations: toolOperationIds } = await import(
	'./webmcpOperations.js'
);

// Each registered WebMCP tool and its operation (webmcpOperations.ts). A tool
// may fold several operations; it may not leave one out.
const toolOperations: Record<string, string[]> = Object.fromEntries(
	Object.entries(toolOperationIds).map(([tool, id]) => [tool, [id]]),
);

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

	test('every operation has a tool or an exclusion reason', async () => {
		const covered = new Set(Object.values(toolOperations).flat());
		const offenders = (await specOperations())
			.filter((op) => !op['x-agent-excluded'])
			.map((op) => op.operationId)
			.filter((id) => !covered.has(id));

		expect(offenders).toEqual([]);
	});

	test('no operation is both a tool and excluded', async () => {
		const covered = new Set(Object.values(toolOperations).flat());
		const both = (await specOperations())
			.filter((op) => 'x-agent-excluded' in op)
			.map((op) => op.operationId)
			.filter((id) => covered.has(id));

		expect(both).toEqual([]);
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
});
