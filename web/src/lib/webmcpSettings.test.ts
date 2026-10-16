import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test';
import { inputSchemas } from './webmcpSchemas.generated.js';
import { registerWebMCPTools } from './webmcpTools.js';

type RegisterToolFn = NonNullable<Document['modelContext']>['registerTool'];
type Tool = Parameters<RegisterToolFn>[0];

const refreshPage = mock(() => Promise.resolve());

function tool(name: string): Tool {
	let found: Tool | undefined;
	registerWebMCPTools(
		{
			registerTool: (registered: Tool) => {
				if (registered.name === name) found = registered;
			},
		},
		{ refreshPage },
	);
	if (!found) throw new Error(`no tool ${name}`);

	return found;
}

const signal = new AbortController().signal;
const originalFetch = globalThis.fetch;
const originalConfirm = globalThis.confirm;

const fetchMock = mock((_url: string, _init?: RequestInit) =>
	Promise.resolve(new Response('{}', { status: 200 })),
);
const confirmMock = mock((_message?: string) => true);

function answer(status: number, body?: unknown) {
	fetchMock.mockImplementationOnce(() =>
		Promise.resolve(
			new Response(body === undefined ? null : JSON.stringify(body), {
				status,
			}),
		),
	);
}

const resolved = { theme: 'dark', defaults: { theme: 'system' } };

beforeEach(() => {
	fetchMock.mockClear();
	confirmMock.mockClear();
	refreshPage.mockClear();
	refreshPage.mockImplementation(() => Promise.resolve());
	globalThis.fetch = fetchMock as unknown as typeof fetch;
	globalThis.confirm = confirmMock;
});

afterEach(() => {
	globalThis.fetch = originalFetch;
	globalThis.confirm = originalConfirm;
});

describe('get_settings', () => {
	test("reads the page's own settings request", async () => {
		answer(200, resolved);

		const result = await tool('get_settings').execute({}, { signal });

		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/settings');
		expect(result.content[0]?.text).toBe(JSON.stringify(resolved));
	});

	test('fails with the status when the server refuses', async () => {
		answer(401, {});

		await expect(tool('get_settings').execute({}, { signal })).rejects.toThrow(
			'401',
		);
	});
});

describe('update_settings', () => {
	test('sends the PATCH the stores send, a null clearing a setting', async () => {
		answer(200, resolved);

		const result = await tool('update_settings').execute(
			{ theme: 'dark', forgeFilter: null },
			{ signal },
		);

		const [url, init] = fetchMock.mock.calls[0] ?? [];
		expect(url).toBe('/api/settings');
		expect(init?.method).toBe('PATCH');
		expect(init?.body).toBe(
			JSON.stringify({ theme: 'dark', forgeFilter: null }),
		);
		expect(result.content[0]?.text).toBe(JSON.stringify(resolved));
	});

	test('asks nobody, as the theme toggle asks nobody', async () => {
		answer(200, resolved);

		await tool('update_settings').execute({ theme: 'light' }, { signal });

		expect(confirmMock).not.toHaveBeenCalled();
	});

	test('has the open page pick the new settings up', async () => {
		answer(200, resolved);

		await tool('update_settings').execute({ theme: 'light' }, { signal });

		expect(refreshPage).toHaveBeenCalledTimes(1);
	});

	test('a page that cannot refresh does not fail the call', async () => {
		answer(200, resolved);
		refreshPage.mockImplementation(() => Promise.reject(new Error('nav')));

		const result = await tool('update_settings').execute(
			{ theme: 'light' },
			{ signal },
		);

		expect(result.content[0]?.text).toBe(JSON.stringify(resolved));
	});

	test("fails with the server's message and does not refresh when it refuses", async () => {
		answer(400, { code: 'bad_request', message: 'theme: invalid' });

		const error = await tool('update_settings')
			.execute({ theme: 'blue' }, { signal })
			.catch((e: unknown) => e as Error);

		expect(error.message).toContain('400');
		expect(error.message).toContain('theme: invalid');
		expect(refreshPage).not.toHaveBeenCalled();
	});

	test('sends a key the schema does not list, so the server decides', async () => {
		answer(400, { message: 'unknown setting' });

		await tool('update_settings')
			.execute({ colour: 'red' }, { signal })
			.catch(() => {});

		expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(
			JSON.stringify({ colour: 'red' }),
		);
	});

	test('works without a refresh hook', async () => {
		answer(200, resolved);
		let found: Tool | undefined;
		registerWebMCPTools({
			registerTool: (registered: Tool) => {
				if (registered.name === 'update_settings') found = registered;
			},
		});

		await found?.execute({ theme: 'light' }, { signal });

		expect(refreshPage).not.toHaveBeenCalled();
	});
});

describe('the update_settings schema', () => {
	const props = inputSchemas.update_settings.properties as Record<
		string,
		{ description?: string; enum?: unknown[]; type?: unknown }
	>;

	test('lists every setting as optional', () => {
		expect(Object.keys(props).sort()).toEqual([
			'forgeFilter',
			'pipelinesHealthFilter',
			'pipelinesRepoSelector',
			'pipelinesSortOrder',
			'telemetryWindow',
			'theme',
		]);
		expect(inputSchemas.update_settings.required).toEqual([]);
	});

	test('lets a closed set take null, which clears it', () => {
		expect(props.theme?.enum).toEqual(['light', 'dark', 'system', null]);
	});

	test.each(Object.keys(inputSchemas.update_settings.properties))(
		'%s says what it controls and that null clears it',
		(key) => {
			expect(props[key]?.description).toMatch(/null/);
		},
	);
});

describe('the settings tools are not marked destructive', () => {
	test.each(['get_settings', 'update_settings'])('%s', (name) => {
		expect(tool(name).annotations).toBeUndefined();
	});
});
