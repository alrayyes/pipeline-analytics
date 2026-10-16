import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test';
import { inputSchemas } from './webmcpSchemas.generated.js';
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

function sentBody(): unknown {
	return JSON.parse(fetchMock.mock.calls[0]?.[1]?.body as string);
}

beforeEach(() => {
	fetchMock.mockClear();
	confirmMock.mockClear();
	confirmMock.mockImplementation(() => true);
	globalThis.fetch = fetchMock as unknown as typeof fetch;
	globalThis.confirm = confirmMock;
});

afterEach(() => {
	globalThis.fetch = originalFetch;
	globalThis.confirm = originalConfirm;
});

describe('list_repos', () => {
	test('asks for 20 repos from offset 0 unless told otherwise', async () => {
		answer(200, { repos: [], hasMore: false });

		const result = await tool('list_repos').execute({}, { signal });

		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/repos?limit=20&offset=0');
		expect(result.content[0]?.text).toBe(
			JSON.stringify({ repos: [], hasMore: false }),
		);
	});

	test('passes the page and the forge through', async () => {
		answer(200, { repos: [], hasMore: false });

		await tool('list_repos').execute(
			{ limit: 5, offset: 10, forge: 'github' },
			{ signal },
		);

		expect(fetchMock.mock.calls[0]?.[0]).toBe(
			'/api/repos?limit=5&offset=10&forge=github',
		);
	});

	test('fails with the status when the server refuses', async () => {
		answer(401, {});

		await expect(tool('list_repos').execute({}, { signal })).rejects.toThrow(
			'401',
		);
	});
});

describe('list_repo_identifiers', () => {
	test('sends the forge and the instance and wraps the identifiers', async () => {
		answer(200, { identifiers: ['a/b'] });

		const result = await tool('list_repo_identifiers').execute(
			{ forge: 'forgejo', forgejoInstanceUrl: 'https://git.example' },
			{ signal },
		);

		expect(fetchMock.mock.calls[0]?.[0]).toBe(
			'/api/repos/identifiers?forge=forgejo&forgejoInstanceUrl=https%3A%2F%2Fgit.example',
		);
		expect(result.content[0]?.text).toBe(
			JSON.stringify({ identifiers: ['a/b'] }),
		);
	});

	test('says forge is required instead of sending "undefined"', async () => {
		await expect(
			tool('list_repo_identifiers').execute({}, { signal }),
		).rejects.toThrow('forge is required');
		expect(fetchMock).not.toHaveBeenCalled();
	});
});

describe('discover_repos', () => {
	test('posts what the Repositories page posts and asks nobody', async () => {
		answer(200, ['a/b', 'c/d']);

		const result = await tool('discover_repos').execute(
			{ forge: 'github', token: 'ghp_secret' },
			{ signal },
		);

		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/repos/discover');
		expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('POST');
		expect(sentBody()).toEqual({ forge: 'github', token: 'ghp_secret' });
		expect(confirmMock).not.toHaveBeenCalled();
		expect(result.content[0]?.text).toBe(
			JSON.stringify({ identifiers: ['a/b', 'c/d'] }),
		);
	});

	test('sends no token when none is given, so the saved one is used', async () => {
		answer(200, []);

		await tool('discover_repos').execute({ forge: 'github' }, { signal });

		expect(sentBody()).toEqual({ forge: 'github' });
	});

	test('says forge is required', async () => {
		await expect(
			tool('discover_repos').execute({}, { signal }),
		).rejects.toThrow('forge is required');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test("fails with the server's own message, which says what to fix", async () => {
		answer(400, { code: 'no_saved_token', message: 'No token saved.' });

		await expect(
			tool('discover_repos').execute({ forge: 'github' }, { signal }),
		).rejects.toThrow('No token saved.');
	});
});

describe('register_repo', () => {
	test('asks the person first, then posts what the page posts', async () => {
		answer(201, { id: 'r1', forge: 'github', identifier: 'a/b' });

		const result = await tool('register_repo').execute(
			{ forge: 'github', identifier: 'a/b', token: 'ghp_secret' },
			{ signal },
		);

		const question = confirmMock.mock.calls[0]?.[0] ?? '';
		expect(question).toContain('a/b');
		expect(question).toContain('github');
		expect(question).not.toContain('ghp_secret');
		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/repos');
		expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('POST');
		expect(sentBody()).toEqual({
			forge: 'github',
			identifier: 'a/b',
			token: 'ghp_secret',
		});
		expect(result.content[0]?.text).toBe(
			JSON.stringify({ id: 'r1', forge: 'github', identifier: 'a/b' }),
		);
		expect(result.content[0]?.text).not.toContain('ghp_secret');
	});

	test('does nothing when the person declines', async () => {
		confirmMock.mockImplementation(() => false);

		await expect(
			tool('register_repo').execute(
				{ forge: 'github', identifier: 'a/b' },
				{ signal },
			),
		).rejects.toThrow('declined');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test('does nothing when the page has no way to ask', async () => {
		// @ts-expect-error -- simulating an environment without confirm()
		globalThis.confirm = undefined;

		await expect(
			tool('register_repo').execute(
				{ forge: 'github', identifier: 'a/b' },
				{ signal },
			),
		).rejects.toThrow('declined');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test('asks through the agent when it offers requestUserInteraction', async () => {
		answer(201, {});
		const requestUserInteraction = mock(<T>(ask: () => Promise<T> | T) =>
			ask(),
		);

		await tool('register_repo').execute(
			{ forge: 'github', identifier: 'a/b' },
			{ signal, requestUserInteraction },
		);

		expect(requestUserInteraction).toHaveBeenCalledTimes(1);
	});

	test.each([
		[{ identifier: 'a/b' }, 'forge'],
		[{ forge: 'github' }, 'identifier'],
	])('says %j is missing %s before asking anyone', async (input, key) => {
		await expect(
			tool('register_repo').execute(input, { signal }),
		).rejects.toThrow(`${key} is required`);
		expect(confirmMock).not.toHaveBeenCalled();
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test("fails with the server's message and code when it refuses", async () => {
		answer(409, { code: 'already_tracked', message: 'Already tracked.' });

		const error = await tool('register_repo')
			.execute({ forge: 'github', identifier: 'a/b' }, { signal })
			.catch((e: unknown) => e as Error);

		expect(error.message).toContain('409');
		expect(error.message).toContain('Already tracked.');
	});
});

describe('untrack_repo', () => {
	test('asks the person first, then sends the DELETE the page sends', async () => {
		answer(204);

		const result = await tool('untrack_repo').execute(
			{ repoId: 'r 1' },
			{ signal },
		);

		const question = confirmMock.mock.calls[0]?.[0] ?? '';
		expect(question).toContain('r 1');
		expect(question).toContain('deleted');
		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/repos/r%201');
		expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('DELETE');
		expect(result.content[0]?.text).toBe(
			JSON.stringify({ repoId: 'r 1', untracked: true }),
		);
	});

	test('does nothing when the person declines', async () => {
		confirmMock.mockImplementation(() => false);

		await expect(
			tool('untrack_repo').execute({ repoId: 'r1' }, { signal }),
		).rejects.toThrow('declined');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test('does nothing when the agent reports the person said no', async () => {
		const requestUserInteraction = mock(<T>(_ask: () => Promise<T> | T) =>
			Promise.resolve(false as T),
		);

		await expect(
			tool('untrack_repo').execute(
				{ repoId: 'r1' },
				{ signal, requestUserInteraction },
			),
		).rejects.toThrow('declined');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	test('says repoId is required before asking anyone', async () => {
		await expect(tool('untrack_repo').execute({}, { signal })).rejects.toThrow(
			'repoId is required',
		);
		expect(confirmMock).not.toHaveBeenCalled();
	});

	test('fails with the status when the repo is gone', async () => {
		answer(404, {});

		await expect(
			tool('untrack_repo').execute({ repoId: 'r1' }, { signal }),
		).rejects.toThrow('404');
	});

	test('tells the agent to pass the id unencoded, as the tool encodes it', () => {
		const description = (
			inputSchemas.untrack_repo.properties.repoId as { description: string }
		).description;

		expect(description).toContain('Do not encode');
	});
});

describe('the write tools are marked destructive', () => {
	test.each(['register_repo', 'untrack_repo'])('%s', (name) => {
		expect(tool(name).annotations).toEqual({ destructiveHint: true });
	});

	test.each(['list_repos', 'list_repo_identifiers', 'discover_repos'])(
		'%s is not',
		(name) => {
			expect(tool(name).annotations).toBeUndefined();
		},
	);
});
