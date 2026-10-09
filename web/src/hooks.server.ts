import type { Handle, ResolveOptions } from '@sveltejs/kit';

// The sheet is inlined into index.html (inlineStyleThreshold, vite.config.ts),
// so the only request left behind it is the font it names. Fetched on
// discovery that is a document -> font chain, which
// network-dependency-tree-insight fails; a rel=preload request is not counted
// as critical, and the browser starts it with the document instead of after
// it. Latin only: the other subsets carry a unicode-range and load when a page
// uses their characters (#527, rules/web-performance.md).
export const preload: NonNullable<ResolveOptions['preload']> = ({
	type,
	path,
}) =>
	type === 'font'
		? path.includes('inter-latin-wght-normal')
		: type === 'js' || type === 'css';

export const handle: Handle = ({ event, resolve }) =>
	resolve(event, { preload });
