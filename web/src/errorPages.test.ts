import { expect, test } from 'bun:test';
import { readFileSync } from 'node:fs';

const read = (path: string): string =>
	readFileSync(new URL(path, import.meta.url), 'utf8');

// src/error.html is SvelteKit's fallback for errors in `handle` and the root
// layout; static/500.html is what the server falls back to. Same wording, one
// source of truth (#576).
test('the SvelteKit fallback error page is the 500 page', () => {
	expect(read('./error.html')).toBe(read('../static/500.html'));
});
