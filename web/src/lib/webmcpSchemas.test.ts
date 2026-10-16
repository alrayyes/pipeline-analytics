import { describe, expect, mock, test } from 'bun:test';
import {
	generateModule,
	loadSpec,
	operationInputSchema,
} from '../../scripts/generate-webmcp-schemas.ts';
import { toolOperations, toolParameterOverrides } from './webmcpOperations.js';

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
}));

const { registerWebMCPTools } = await import('./webmcpTools.js');

type Tool = {
	name: string;
	description: string;
	inputSchema?: unknown;
};

function registeredTools(): Tool[] {
	const tools: Tool[] = [];
	registerWebMCPTools({
		registerTool: (tool: Tool) => {
			tools.push(tool);
		},
	} as unknown as NonNullable<Document['modelContext']>);

	return tools;
}

// Stryker copies only web/ into its sandbox, so the spec one level above is not
// there (see webmcpParity.test.ts). There is no mutated source behind these.
const inStryker = process.env.STRYKER_MUTATOR_WORKER !== undefined;

describe.skipIf(inStryker)('WebMCP input schemas follow openapi.yaml', () => {
	test('every tool declares the schema generated from its operation', async () => {
		const spec = await loadSpec();

		const mismatched = registeredTools()
			.filter(
				(tool) =>
					JSON.stringify(tool.inputSchema) !==
					JSON.stringify(
						operationInputSchema(
							spec,
							toolOperations[tool.name] as string,
							toolParameterOverrides[tool.name],
						),
					),
			)
			.map((tool) => tool.name);

		expect(mismatched).toEqual([]);
	});

	test('the committed generated module is current, so a changed spec parameter fails here', async () => {
		const committed = await Bun.file(
			new URL('./webmcpSchemas.generated.ts', import.meta.url),
		).text();

		expect(committed).toBe(
			generateModule(await loadSpec(), toolOperations, toolParameterOverrides),
		);
	});
});

describe('WebMCP tool descriptions', () => {
	test('list_runs and list_pipelines state their default paging', () => {
		const tools = registeredTools();

		for (const name of ['list_runs', 'list_pipelines']) {
			expect(tools.find((tool) => tool.name === name)?.description).toContain(
				'20',
			);
		}
	});
});
