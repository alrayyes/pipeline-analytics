// Bun has no official StrykerJS runner plugin (javascript.md). Of the two
// community ones, @hughescr/stryker-bun-runner is the one actually alive:
// pushed within the last week, on @stryker-mutator/core ^10 (this repo's
// pinned version), versus stryker-mutator-bun-runner's last publish over a
// year ago, stuck on an older core range.
export default {
	plugins: ['@hughescr/stryker-bun-runner'],
	testRunner: 'bun',
	coverageAnalysis: 'perTest',
	// Scoped to the modules the unit-test layer (#228) actually covers --
	// mutating a .svelte component or a route file Stryker can't run
	// bun:test against would just report every mutant as uncovered noise.
	mutate: [
		'src/lib/format.ts',
		'src/lib/flakyRuns.ts',
		'src/lib/relativeTime.ts',
		'src/lib/pipelinesFilters.svelte.ts',
		'src/lib/forgeFilter.svelte.ts',
	],
	concurrency: 4,
	reporters: ['clear-text', 'html'],
};
