import { describe, expect, test } from 'bun:test';
import { parseAnsiLines, type Segment } from './ansi.js';

const text = (line: Segment[]) => line.map((s) => s.text).join('');

describe('parseAnsiLines', () => {
	test('a plain line is one unstyled segment', () => {
		expect(parseAnsiLines(['hello'])).toEqual([[{ text: 'hello' }]]);
	});

	test('a colour applies until it is reset', () => {
		expect(parseAnsiLines(['\x1b[31mred\x1b[0m plain'])).toEqual([
			[{ text: 'red', fg: 'red' }, { text: ' plain' }],
		]);
	});

	test('bright colours and bold are kept apart from the plain ones', () => {
		expect(parseAnsiLines(['\x1b[1;92mok\x1b[22m still green'])).toEqual([
			[
				{ text: 'ok', fg: 'bright-green', bold: true },
				{ text: ' still green', fg: 'bright-green' },
			],
		]);
	});

	test('a colour carries onto the next line until reset, as in a terminal', () => {
		expect(parseAnsiLines(['\x1b[33mwarn', 'more\x1b[39m', 'done'])).toEqual([
			[{ text: 'warn', fg: 'yellow' }],
			[{ text: 'more', fg: 'yellow' }],
			[{ text: 'done' }],
		]);
	});

	test('256-colour and truecolour sequences are skipped without eating the text', () => {
		const [line] = parseAnsiLines(['\x1b[38;5;196mfoo\x1b[38;2;1;2;3mbar']);

		expect(text(line ?? [])).toBe('foobar');
	});

	test('cursor movement, OSC titles and stray escapes are dropped', () => {
		const [line] = parseAnsiLines([
			'a\x1b[2Kb\x1b]0;title\x07c\x1b[?25ld\x1bZe',
		]);

		expect(text(line ?? [])).toBe('abcde');
	});

	test('control characters are removed but a tab survives', () => {
		const [line] = parseAnsiLines(['a\r\x00\x07b\tc']);

		expect(text(line ?? [])).toBe('ab\tc');
	});

	test('markup stays text: a script tag is never interpreted', () => {
		const payload = '<script>alert(1)</script><img src=x onerror=alert(2)>';
		const [line] = parseAnsiLines([`\x1b[31m${payload}\x1b[0m`]);

		expect(line).toEqual([{ text: payload, fg: 'red' }]);
	});

	test('an empty line stays an empty line', () => {
		expect(parseAnsiLines(['', 'x'])).toEqual([[], [{ text: 'x' }]]);
	});
});
