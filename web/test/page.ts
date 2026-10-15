import { flushSync } from 'svelte';

// Shared by the page component tests (*.dom.test.ts): a fake network and a few
// helpers to drive a mounted page. Import test/dom.ts first in the test file.

export function json(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'content-type': 'application/json' },
	});
}

// Lets a mounted page's effects run and its fetches settle.
export async function settle(): Promise<void> {
	for (let i = 0; i < 5; i++) {
		await Promise.resolve();
		await new Promise((resolve) => setTimeout(resolve, 0));
		flushSync();
	}
}

// A radio in a toggle group, found by its visible name, as a user would.
export function radio(name: string): HTMLElement {
	const found = [
		...document.querySelectorAll<HTMLElement>('[role="radio"]'),
	].find((el) => el.textContent?.trim() === name);
	if (!found) throw new Error(`no radio named ${name}`);

	return found;
}

// Replaces fetch with a fake that records every URL it is asked for and answers
// from `answer`; a URL it returns nothing for gets an empty object. `restore`
// puts the real one back.
export function fakeNetwork(answer: (url: URL) => Response | undefined) {
	const requested: URL[] = [];
	const real = globalThis.fetch;

	globalThis.fetch = (async (input: RequestInfo | URL) => {
		const url = new URL(String(input), 'http://localhost');
		requested.push(url);

		return answer(url) ?? json({});
	}) as typeof fetch;

	return {
		requested,
		restore: () => {
			globalThis.fetch = real;
		},
	};
}
