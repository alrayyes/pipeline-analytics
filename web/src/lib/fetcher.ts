// What the API modules need from `fetch`: the call, not the rest of its
// namespace. `typeof fetch` also demands `preconnect` under Bun's types, which
// a wrapper or a test double has no reason to carry.
export type Fetcher = (
	input: RequestInfo | URL,
	init?: RequestInit,
) => Promise<Response>;
