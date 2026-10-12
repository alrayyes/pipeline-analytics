import { expect, test } from 'bun:test';
import { readFileSync } from 'node:fs';

const read = (path: string): string =>
	readFileSync(new URL(path, import.meta.url), 'utf8');

const NOINDEX = '<meta name="robots" content="noindex, nofollow" />';

// An internal tool must not be indexed (#589). Every HTML document the server
// can answer with carries the tag, so a crawler that skips the X-Robots-Tag
// header still reads it.
for (const page of [
	'./app.html',
	'./error.html',
	'../static/404.html',
	'../static/500.html',
]) {
	test(`${page.split('/').pop()} tells search engines not to index it`, () => {
		expect(read(page)).toContain(NOINDEX);
	});
}

// The tag is only readable if the crawler may fetch the page (#589).
test('robots.txt does not disallow crawling', () => {
	let robots = '';
	try {
		robots = read('../static/robots.txt');
	} catch {
		// absent is fine
	}
	expect(robots).not.toMatch(/^\s*Disallow:\s*\/\s*$/m);
});
