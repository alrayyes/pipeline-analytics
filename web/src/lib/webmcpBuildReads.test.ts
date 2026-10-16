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
const originalFetch = globalThis.fetch;
const fetchMock = mock((_url: string, _init?: RequestInit) =>
	Promise.resolve(new Response('{}', { status: 200 })),
);

function answer(status: number, body?: unknown) {
	fetchMock.mockImplementationOnce(() =>
		Promise.resolve(
			new Response(body === undefined ? null : JSON.stringify(body), {
				status,
			}),
		),
	);
}

beforeEach(() => {
	fetchMock.mockClear();
	globalThis.fetch = fetchMock as unknown as typeof fetch;
});

afterEach(() => {
	globalThis.fetch = originalFetch;
});

describe('get_version', () => {
	test("reads the footer's own version request", async () => {
		answer(200, { version: 'v1.2.3' });

		const result = await tool('get_version').execute({}, { signal });

		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/version');
		expect(result.content[0]?.text).toBe(JSON.stringify({ version: 'v1.2.3' }));
	});

	test('fails with the status when the server refuses', async () => {
		answer(503, {});

		await expect(tool('get_version').execute({}, { signal })).rejects.toThrow(
			'503',
		);
	});
});

describe('get_github_rate_limit', () => {
	test("reads the Insights page's own request and wraps the list", async () => {
		const tokens = [{ tokenMasked: '****1234', repos: ['a/b'] }];
		answer(200, tokens);

		const result = await tool('get_github_rate_limit').execute({}, { signal });

		expect(fetchMock.mock.calls[0]?.[0]).toBe(
			'/api/insights/github-rate-limit',
		);
		expect(result.content[0]?.text).toBe(JSON.stringify({ tokens }));
	});

	test('fails with the status when the server refuses', async () => {
		answer(401, {});

		await expect(
			tool('get_github_rate_limit').execute({}, { signal }),
		).rejects.toThrow('401');
	});
});
