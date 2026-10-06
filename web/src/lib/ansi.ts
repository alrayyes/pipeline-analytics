// Turns raw CI log lines into styled text, for the step log viewer (#343).
// The output is segments of plain text plus a colour name, never HTML, so a
// log that contains markup renders as that markup: the component draws each
// segment through text interpolation and nothing here ever builds a string
// that is parsed as HTML.

export type AnsiColour =
	| 'black'
	| 'red'
	| 'green'
	| 'yellow'
	| 'blue'
	| 'magenta'
	| 'cyan'
	| 'white'
	| 'bright-black'
	| 'bright-red'
	| 'bright-green'
	| 'bright-yellow'
	| 'bright-blue'
	| 'bright-magenta'
	| 'bright-cyan'
	| 'bright-white';

export interface Segment {
	text: string;
	fg?: AnsiColour;
	bold?: true;
}

const BASE: AnsiColour[] = [
	'black',
	'red',
	'green',
	'yellow',
	'blue',
	'magenta',
	'cyan',
	'white',
];

const BRIGHT: AnsiColour[] = BASE.map((name) => `bright-${name}` as AnsiColour);

interface Style {
	fg?: AnsiColour;
	bold: boolean;
}

// ESC [ params final-byte, ESC ] ... (BEL | ESC \), and any other lone ESC
// with the one character after it.
const ESCAPE =
	// biome-ignore lint/suspicious/noControlCharactersInRegex: matching escape sequences is the point.
	/\x1b\[([0-9;:?<=>]*)([\x20-\x2f]*[\x40-\x7e])|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)?|\x1b./gs;

// Everything below space except a tab, plus DEL.
// biome-ignore lint/suspicious/noControlCharactersInRegex: stripping them is the point.
const CONTROL = /[\x00-\x08\x0a-\x1f\x7f]/g;

function applySgr(style: Style, params: string): Style {
	const next = { ...style };
	const codes = params === '' ? [0] : params.split(/[;:]/).map(Number);

	for (let i = 0; i < codes.length; i++) {
		const code = codes[i] ?? 0;

		if (code === 0) {
			next.fg = undefined;
			next.bold = false;
		} else if (code === 1) {
			next.bold = true;
		} else if (code === 22) {
			next.bold = false;
		} else if (code >= 30 && code <= 37) {
			next.fg = BASE[code - 30];
		} else if (code === 39) {
			next.fg = undefined;
		} else if (code >= 90 && code <= 97) {
			next.fg = BRIGHT[code - 90];
		} else if (code === 38 || code === 48) {
			// 256-colour (5;n) and truecolour (2;r;g;b): skip their operands
			// so they aren't read as separate codes. Not mapped to a colour.
			i += codes[i + 1] === 5 ? 2 : codes[i + 1] === 2 ? 4 : 0;
		}
	}

	return next;
}

function parseLine(line: string, start: Style): [Segment[], Style] {
	const segments: Segment[] = [];
	let style = start;
	let last = 0;

	const push = (raw: string) => {
		const text = raw.replace(CONTROL, '');
		if (text === '') return;

		const segment: Segment = { text };
		if (style.fg) segment.fg = style.fg;
		if (style.bold) segment.bold = true;
		segments.push(segment);
	};

	for (const match of line.matchAll(ESCAPE)) {
		push(line.slice(last, match.index));
		last = match.index + match[0].length;

		if (match[2] === 'm') style = applySgr(style, match[1] ?? '');
	}
	push(line.slice(last));

	return [segments, style];
}

// A colour set on one line carries onto the next until it is reset, as it
// does in a terminal, so state threads through the lines in order.
export function parseAnsiLines(lines: string[]): Segment[][] {
	let style: Style = { bold: false };

	return lines.map((line) => {
		const [segments, next] = parseLine(line, style);
		style = next;

		return segments;
	});
}
