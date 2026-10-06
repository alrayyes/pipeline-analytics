import { describe, expect, test } from 'bun:test';
import { actionMessage, offersRerun, rerunRun } from './runActions.js';

function errorResponse(status: number, code: string): Response {
	return new Response(JSON.stringify({ code, message: 'server text' }), {
		status,
	});
}

describe('rerunRun', () => {
	test('posts to the run’s rerun endpoint and accepts a 202', async () => {
		let url: string | undefined;
		let method: string | undefined;
		const fetchFn = (input: RequestInfo | URL, init?: RequestInit) => {
			url = String(input);
			method = init?.method;

			return Promise.resolve(new Response(null, { status: 202 }));
		};

		const result = await rerunRun('run 1', fetchFn as unknown as typeof fetch);

		expect(url).toBe('/api/runs/run%201/rerun');
		expect(method).toBe('POST');
		expect(result).toEqual({ ok: true });
	});

	test.each([
		[403, 'forbidden', 'token needs Actions write permission'],
		[409, 'not_actionable', 'can only be re-run once it has finished'],
		[501, 'unsupported', 'doesn’t support re-running'],
		[502, 'unreachable', 'forge didn’t answer'],
		[404, 'not_found', 'Run not found'],
	])('maps %d %s to its own message', async (status, code, fragment) => {
		const fetchFn = () => Promise.resolve(errorResponse(status, code));

		const result = await rerunRun('r', fetchFn as unknown as typeof fetch);

		if (result.ok) throw new Error('expected a failure');
		expect(actionMessage('rerun', result)).toContain(fragment);
	});

	test('a network failure reads as the server being unreachable', async () => {
		const fetchFn = () => Promise.reject(new TypeError('offline'));

		const result = await rerunRun('r', fetchFn as unknown as typeof fetch);

		expect(result).toEqual({ ok: false, status: 0, code: 'network' });
		if (result.ok) throw new Error('expected a failure');
		expect(actionMessage('rerun', result)).toContain('Could not reach');
	});
});

describe('offersRerun', () => {
	test('is true only when the server lists rerun', () => {
		expect(offersRerun({ actions: ['rerun'] })).toBe(true);
		expect(offersRerun({ actions: ['cancel'] })).toBe(false);
		expect(offersRerun({ actions: [] })).toBe(false);
	});
});
