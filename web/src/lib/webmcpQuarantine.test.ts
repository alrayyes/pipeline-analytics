import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test';
import { registerWebMCPTools } from './webmcpTools.js';

type RegisterToolFn = NonNullable<Document['modelContext']>['registerTool'];
type Tool = Parameters<RegisterToolFn>[0];

function tool(name: string): Tool {
	let found: Tool | undefined;
	registerWebMCPTools({
		registerTool: (registered: Tool) => {
			if (registered.name === name) found = registered;
		},
	});
	if (!found) throw new Error(`no tool ${name}`);

	return found;
}

const signal = new AbortController().signal;
const original = globalThis.fetch;
const fetchMock = mock((_url: string, _init?: RequestInit) =>
	Promise.resolve(
		new Response(JSON.stringify({ quarantined: true }), { status: 200 }),
	),
);

beforeEach(() => {
	fetchMock.mockClear();
	globalThis.fetch = fetchMock as unknown as typeof fetch;
});

afterEach(() => {
	globalThis.fetch = original;
});

// The tools call the page's own quarantineStep and unquarantineStep, so what
// is pinned here is the request the Flaky page would send.
describe('quarantine_step', () => {
	test('sends the PUT the Flaky page sends, with the step escaped', async () => {
		const result = await tool('quarantine_step').execute(
			{ pipelineId: 'p1', step: 'lint/css', note: 'known flake' },
			{ signal },
		);

		const [url, init] = fetchMock.mock.calls[0] ?? [];
		expect(url).toBe('/api/pipelines/p1/steps/lint%2Fcss/quarantine');
		expect(init?.method).toBe('PUT');
		expect(init?.body).toBe(JSON.stringify({ note: 'known flake' }));
		expect(result.content[0]?.text).toBe(JSON.stringify({ quarantined: true }));
	});

	test('sends no body when there is no note', async () => {
		await tool('quarantine_step').execute(
			{ pipelineId: 'p1', step: 'unit' },
			{ signal },
		);

		expect(fetchMock.mock.calls[0]?.[1]?.body).toBeUndefined();
	});

	test.each([
		[{ step: 'unit' }, 'pipelineId'],
		[{ pipelineId: 'p1' }, 'step'],
	])(
		'says %j is missing %s instead of sending "undefined"',
		async (input, key) => {
			await expect(
				tool('quarantine_step').execute(input, { signal }),
			).rejects.toThrow(`${key} is required`);
			expect(fetchMock).not.toHaveBeenCalled();
		},
	);

	test("fails with the server's status when it refuses", async () => {
		fetchMock.mockImplementationOnce(() =>
			Promise.resolve(new Response('{}', { status: 422 })),
		);

		await expect(
			tool('quarantine_step').execute(
				{ pipelineId: 'p1', step: 'unit', note: 'x' },
				{ signal },
			),
		).rejects.toThrow('422');
	});
});

describe('unquarantine_step', () => {
	test('sends the DELETE the Flaky page sends', async () => {
		await tool('unquarantine_step').execute(
			{ pipelineId: 'p1', step: 'unit' },
			{ signal },
		);

		const [url, init] = fetchMock.mock.calls[0] ?? [];
		expect(url).toBe('/api/pipelines/p1/steps/unit/quarantine');
		expect(init?.method).toBe('DELETE');
	});

	test('says step is required instead of sending "undefined"', async () => {
		await expect(
			tool('unquarantine_step').execute({ pipelineId: 'p1' }, { signal }),
		).rejects.toThrow('step is required');
	});
});

describe('quarantine tool schemas', () => {
	test('tell an agent to pass the step name as listed, since the tool encodes it', async () => {
		const { inputSchemas } = await import('./webmcpSchemas.generated.js');

		for (const tool of ['quarantine_step', 'unquarantine_step'] as const) {
			const description = inputSchemas[tool].properties.step.description;

			expect(description).not.toContain('percent-encoded');
			expect(description).toContain('Do not encode');
		}
	});
});
