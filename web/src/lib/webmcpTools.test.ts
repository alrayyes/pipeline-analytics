import { beforeEach, describe, expect, mock, test } from 'bun:test';

const fetchPipelinesMock = mock(() =>
	Promise.resolve({ pipelines: [], hasMore: false }),
);
const fetchPipelineMock = mock(() => Promise.resolve({ id: 'p1' }));
const fetchRepoUsageMock = mock(() =>
	Promise.resolve([{ workflow: 'CI', runnerMinutes: 1 }]),
);

const fetchPipelineStepsMock = mock(() => Promise.resolve([{ name: 'unit' }]));
const fetchJobLogMock = mock(() => Promise.resolve({ available: true }));
const fetchFailureInsightsMock = mock((_params: { window?: string }) =>
	Promise.resolve({ window: '7d' }),
);
const fetchRunsMock = mock(() => Promise.resolve({ runs: [], hasMore: false }));
const fetchFlakyStepsMock = mock(
	(_params: { limit?: number; offset?: number }) =>
		Promise.resolve({ steps: [] }),
);
const fetchBranchesMock = mock(() => Promise.resolve({ branches: [] }));
const fetchRunStepsMock = mock((_runId: string) =>
	Promise.resolve({ steps: [] }),
);
const fetchFlakyRunsMock = mock(
	(_pipelineId: string, _params: { step: string; window?: string }) =>
		Promise.resolve([]),
);
const fetchUnhealthyStepsMock = mock(
	(_params: { window?: string; limit?: number; offset?: number }) =>
		Promise.resolve({ groups: [], hasMore: false }),
);

mock.module('./dashboardApi.js', () => ({
	fetchPipelines: fetchPipelinesMock,
	fetchPipeline: fetchPipelineMock,
	fetchRepoUsage: fetchRepoUsageMock,
	fetchPipelineSteps: fetchPipelineStepsMock,
	fetchJobLog: fetchJobLogMock,
	fetchFailureInsights: fetchFailureInsightsMock,
	fetchRuns: fetchRunsMock,
	fetchFlakySteps: fetchFlakyStepsMock,
	fetchBranches: fetchBranchesMock,
	fetchRunSteps: fetchRunStepsMock,
	fetchFlakyRuns: fetchFlakyRunsMock,
	fetchUnhealthySteps: fetchUnhealthyStepsMock,
}));

const { registerWebMCPTools } = await import('./webmcpTools.js');

type RegisterToolFn = NonNullable<Document['modelContext']>['registerTool'];
type Tool = Parameters<RegisterToolFn>[0];

function newRegisterToolMock() {
	return mock((_tool: Tool, _options?: Parameters<RegisterToolFn>[1]) => {});
}

beforeEach(() => {
	fetchPipelinesMock.mockClear();
	fetchPipelineMock.mockClear();
	fetchRepoUsageMock.mockClear();
	fetchPipelineStepsMock.mockClear();
	fetchJobLogMock.mockClear();
	fetchFailureInsightsMock.mockClear();
	fetchRunsMock.mockClear();
	fetchFlakyStepsMock.mockClear();
	fetchBranchesMock.mockClear();
	fetchRunStepsMock.mockClear();
	fetchFlakyRunsMock.mockClear();
	fetchUnhealthyStepsMock.mockClear();
});

function findTool(
	registerTool: ReturnType<typeof newRegisterToolMock>,
	name: string,
): Tool {
	const call = registerTool.mock.calls.find(([tool]) => tool.name === name);
	if (!call) throw new Error(`no tool registered with name ${name}`);
	return call[0];
}

describe('registerWebMCPTools', () => {
	test('does not throw and registers nothing when modelContext is undefined', () => {
		expect(() => registerWebMCPTools(undefined)).not.toThrow();
	});

	test('registers each tool under the name the MCP server uses', () => {
		const registerTool = newRegisterToolMock();

		registerWebMCPTools({ registerTool });

		const names = registerTool.mock.calls.map(([tool]) => tool.name);
		expect(names).toEqual([
			'list_pipelines',
			'get_pipeline',
			'get_repo_usage',
			'list_pipeline_steps',
			'get_job_log',
			'get_failure_insights',
			'list_runs',
			'list_flaky_steps',
			'list_branches',
			'get_run_steps',
			'list_pipeline_flaky_runs',
			'list_unhealthy_steps',
		]);
	});

	test("list_pipelines' execute calls fetchPipelines and returns its result as the tool's content", async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		const tool = findTool(registerTool, 'list_pipelines');
		const result = await tool.execute(
			{},
			{ signal: new AbortController().signal },
		);

		expect(fetchPipelinesMock).toHaveBeenCalledTimes(1);
		expect(result).toEqual({
			content: [
				{
					type: 'text',
					text: JSON.stringify({ pipelines: [], hasMore: false }),
				},
			],
		});
	});

	test("get_pipeline's execute calls fetchPipeline with the pipelineId argument it's given", async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		const tool = findTool(registerTool, 'get_pipeline');
		const result = await tool.execute(
			{ pipelineId: 'p1' },
			{ signal: new AbortController().signal },
		);

		expect(fetchPipelineMock).toHaveBeenCalledWith('p1');
		expect(result).toEqual({
			content: [{ type: 'text', text: JSON.stringify({ id: 'p1' }) }],
		});
	});

	test("get_repo_usage's execute calls fetchRepoUsage with the repoId argument it's given", async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		const tool = findTool(registerTool, 'get_repo_usage');
		const result = await tool.execute(
			{ repoId: 'r1' },
			{ signal: new AbortController().signal },
		);

		expect(fetchRepoUsageMock).toHaveBeenCalledWith('r1');
		expect(result).toEqual({
			content: [
				{
					type: 'text',
					text: JSON.stringify([{ workflow: 'CI', runnerMinutes: 1 }]),
				},
			],
		});
	});
});

// What each read tool hands its fetch function. The tool's own logic is only
// this mapping, so one call per tool pins it; the fetch functions and the
// endpoints are tested where they live.
describe('read tools', () => {
	const cases: [
		string,
		Record<string, unknown>,
		ReturnType<typeof mock>,
		unknown[],
	][] = [
		[
			'list_pipeline_steps',
			{ pipelineId: 'p1', window: '30d' },
			fetchPipelineStepsMock,
			['p1', { window: '30d' }],
		],
		[
			'get_job_log',
			{ runId: 'r1', jobId: 'j1', lines: 500 },
			fetchJobLogMock,
			['r1', 'j1', { lines: 500 }],
		],
		[
			'get_failure_insights',
			{ window: '30d', repoId: 'r1', forge: 'github', branch: 'main' },
			fetchFailureInsightsMock,
			[{ window: '30d', repoId: 'r1', forge: 'github', branch: 'main' }],
		],
		[
			'list_runs',
			{ status: 'failed', limit: 5, offset: 10, branch: 'main' },
			fetchRunsMock,
			[{ status: 'failed', limit: 5, offset: 10, branch: 'main' }],
		],
		[
			'list_flaky_steps',
			{ window: '24h', limit: 3 },
			fetchFlakyStepsMock,
			[{ window: '24h', limit: 3 }],
		],
		['list_branches', { window: '7d' }, fetchBranchesMock, [{ window: '7d' }]],
		['get_run_steps', { runId: 'r1' }, fetchRunStepsMock, ['r1']],
		[
			'list_pipeline_flaky_runs',
			{ pipelineId: 'p1', step: 'unit', window: '30d' },
			fetchFlakyRunsMock,
			['p1', { step: 'unit', window: '30d' }],
		],
		[
			'list_unhealthy_steps',
			{ window: '24h', limit: 5, offset: 10 },
			fetchUnhealthyStepsMock,
			[{ window: '24h', limit: 5, offset: 10 }],
		],
	];

	for (const [name, input, fetchMock, args] of cases) {
		test(`${name} calls its fetch function with the arguments it is given`, async () => {
			const registerTool = newRegisterToolMock();
			registerWebMCPTools({ registerTool });

			const result = await findTool(registerTool, name).execute(input, {
				signal: new AbortController().signal,
			});

			expect(fetchMock).toHaveBeenCalledWith(...args);
			expect(result.content[0].type).toBe('text');
		});
	}

	test('list_runs asks for the first 20 when no paging is given', async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		await findTool(registerTool, 'list_runs').execute(
			{},
			{ signal: new AbortController().signal },
		);

		expect(fetchRunsMock).toHaveBeenCalledWith({ limit: 20, offset: 0 });
	});

	test('list_runs ignores a paging value that is not a number', async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		await findTool(registerTool, 'list_runs').execute(
			{ limit: '5', offset: null },
			{ signal: new AbortController().signal },
		);

		expect(fetchRunsMock).toHaveBeenCalledWith({ limit: 20, offset: 0 });
	});

	test('list_flaky_steps sends no paging when none is a number', async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		await findTool(registerTool, 'list_flaky_steps').execute(
			{ limit: '3', offset: {} },
			{ signal: new AbortController().signal },
		);

		const sent = fetchFlakyStepsMock.mock.calls[0]?.[0];
		expect([sent?.limit, sent?.offset]).toEqual([undefined, undefined]);
	});

	test('get_failure_insights invents no window, so the server keeps its default', async () => {
		const registerTool = newRegisterToolMock();
		registerWebMCPTools({ registerTool });

		await findTool(registerTool, 'get_failure_insights').execute(
			{},
			{ signal: new AbortController().signal },
		);

		expect(fetchFailureInsightsMock.mock.calls[0]?.[0]?.window).toBeUndefined();
	});
});
