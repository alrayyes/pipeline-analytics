import { beforeEach, describe, expect, mock, test } from 'bun:test';

const fetchPipelinesMock = mock(() =>
	Promise.resolve({ pipelines: [], hasMore: false }),
);
const fetchPipelineMock = mock(() => Promise.resolve({ id: 'p1' }));
const fetchRepoUsageMock = mock(() =>
	Promise.resolve([{ workflow: 'CI', runnerMinutes: 1 }]),
);

mock.module('./dashboardApi.js', () => ({
	fetchPipelines: fetchPipelinesMock,
	fetchPipeline: fetchPipelineMock,
	fetchRepoUsage: fetchRepoUsageMock,
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

	test('registers exactly three tools, named list_pipelines, get_pipeline, get_repo_usage', () => {
		const registerTool = newRegisterToolMock();

		registerWebMCPTools({ registerTool });

		expect(registerTool).toHaveBeenCalledTimes(3);
		const names = registerTool.mock.calls.map(([tool]) => tool.name);
		expect(names).toEqual(['list_pipelines', 'get_pipeline', 'get_repo_usage']);
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
