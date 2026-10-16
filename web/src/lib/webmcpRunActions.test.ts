import { afterEach, beforeEach, describe, expect, mock, test } from 'bun:test';

const rerunRunMock = mock((_runId: string) =>
	Promise.resolve<{ ok: true } | { ok: false; status: number; code: string }>({
		ok: true,
	}),
);
const cancelRunMock = mock((_runId: string) =>
	Promise.resolve<{ ok: true } | { ok: false; status: number; code: string }>({
		ok: true,
	}),
);

mock.module('./runActions.js', () => ({
	rerunRun: rerunRunMock,
	cancelRun: cancelRunMock,
	actionMessage: (kind: string, result: { code: string }) =>
		`${kind} refused: ${result.code}`,
}));

const { registerWebMCPTools } = await import('./webmcpTools.js');

type RegisterToolFn = NonNullable<Document['modelContext']>['registerTool'];
type Tool = Parameters<RegisterToolFn>[0];

function tools(): Map<string, Tool> {
	const registered = new Map<string, Tool>();
	registerWebMCPTools({
		registerTool: (tool: Tool) => {
			registered.set(tool.name, tool);
		},
	});

	return registered;
}

const signal = new AbortController().signal;
const original = globalThis.confirm;
const confirmMock = mock((_message?: string) => true);

beforeEach(() => {
	rerunRunMock.mockClear();
	cancelRunMock.mockClear();
	confirmMock.mockClear();
	confirmMock.mockImplementation(() => true);
	globalThis.confirm = confirmMock;
});

afterEach(() => {
	globalThis.confirm = original;
});

describe.each([
	['rerun_run', rerunRunMock, 're-run'],
	['cancel_run', cancelRunMock, 'cancel'],
] as const)('%s', (name, actionMock, verb) => {
	test('asks the person first, then calls the same action the Runs page does', async () => {
		const result = await tools()
			.get(name)
			?.execute({ runId: 'r1' }, { signal });

		expect(confirmMock.mock.calls[0]?.[0]).toContain(verb);
		expect(confirmMock.mock.calls[0]?.[0]).toContain('r1');
		expect(actionMock).toHaveBeenCalledWith('r1');
		expect(result?.content[0]?.text).toContain('"accepted":true');
	});

	test('does nothing when the person declines', async () => {
		confirmMock.mockImplementation(() => false);

		await expect(
			tools().get(name)?.execute({ runId: 'r1' }, { signal }),
		).rejects.toThrow('declined');
		expect(actionMock).not.toHaveBeenCalled();
	});

	test('does nothing when the page has no way to ask', async () => {
		// @ts-expect-error -- simulating an environment without confirm()
		globalThis.confirm = undefined;

		await expect(
			tools().get(name)?.execute({ runId: 'r1' }, { signal }),
		).rejects.toThrow('declined');
		expect(actionMock).not.toHaveBeenCalled();
	});

	test('asks through the agent when it offers requestUserInteraction', async () => {
		const requestUserInteraction = mock(<T>(ask: () => Promise<T> | T) =>
			ask(),
		);

		await tools()
			.get(name)
			?.execute({ runId: 'r1' }, { signal, requestUserInteraction });

		expect(requestUserInteraction).toHaveBeenCalledTimes(1);
		expect(actionMock).toHaveBeenCalledWith('r1');
	});

	test('says runId is required instead of acting on "undefined"', async () => {
		await expect(tools().get(name)?.execute({}, { signal })).rejects.toThrow(
			'runId is required',
		);
		expect(confirmMock).not.toHaveBeenCalled();
	});

	test("raises the server's refusal in the page's own words", async () => {
		actionMock.mockImplementationOnce(() =>
			Promise.resolve({ ok: false, status: 403, code: 'forbidden' }),
		);

		await expect(
			tools().get(name)?.execute({ runId: 'r1' }, { signal }),
		).rejects.toThrow('refused: forbidden');
	});

	test('is marked destructive so an agent host can gate it too', () => {
		expect(tools().get(name)?.annotations).toMatchObject({
			destructiveHint: true,
		});
	});
});
