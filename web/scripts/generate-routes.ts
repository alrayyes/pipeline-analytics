// Writes static/routes.json: the dashboard's routes, as patterns like
// `/runs/[id]`, read from src/routes. The Go server loads it from the embedded
// dist/ to tell a client route from a mistyped URL, and answers 404 for the
// latter instead of serving the app shell as 200 (#577). Runs as `prebuild`
// beside generate-releases.ts, and the output is gitignored for the same reason.
//
// It reads the tree SvelteKit's router reads, so a new route reaches the server
// without a second list to keep.

import { mkdirSync, readdirSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

// A directory is a route when it holds a page: `+page.svelte`, or just a
// `+page.ts` that redirects. Layouts, error views and `+server` handlers
// don't add one.
const PAGE = /^\+page(@.*)?\.(svelte|ts|js)$/;

export function routePatterns(routesDir: string, prefix = ''): string[] {
	const entries = readdirSync(routesDir);
	const patterns: string[] = [];

	if (entries.some((entry) => PAGE.test(entry))) {
		patterns.push(prefix === '' ? '/' : prefix);
	}

	for (const entry of entries) {
		if (statSync(join(routesDir, entry)).isDirectory()) {
			patterns.push(
				...routePatterns(join(routesDir, entry), `${prefix}/${entry}`),
			);
		}
	}

	return patterns.sort();
}

if (import.meta.main) {
	const output = join(import.meta.dir, '..', 'static', 'routes.json');
	const patterns = routePatterns(join(import.meta.dir, '..', 'src', 'routes'));

	mkdirSync(dirname(output), { recursive: true });
	writeFileSync(output, `${JSON.stringify(patterns, null, '\t')}\n`);

	console.log(`routes.json: ${patterns.length} routes`);
}
