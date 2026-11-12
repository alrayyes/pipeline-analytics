import { mock } from 'bun:test';

// The agent's `requestUserInteraction`, as a mock a test can count calls on.
// Its real type is generic in the answer, which `mock()` can't carry over, so
// the one cast lives here instead of at every call site.
export type UserInteraction = <T>(ask: () => Promise<T> | T) => Promise<T> | T;

export function userInteraction(respond: (ask: () => unknown) => unknown) {
	return mock(respond) as unknown as UserInteraction &
		ReturnType<typeof mock<typeof respond>>;
}
