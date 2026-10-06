// Bun has no official StrykerJS runner plugin (javascript.md). Of the two
// community ones, @hughescr/stryker-bun-runner is the one actually alive:
// pushed within the last week, on @stryker-mutator/core ^10 (this repo's
// pinned version), versus stryker-mutator-bun-runner's last publish over a
// year ago, stuck on an older core range.
export default {
	plugins: ['@hughescr/stryker-bun-runner'],
	testRunner: 'bun',
	// Same flag as `bun run test`: each file gets its own module registry, so
	// a test that reads a store's initial value can't see another file's writes.
	bun: { bunArgs: ['--isolate'] },
	coverageAnalysis: 'perTest',
	// Stryker's sandbox is built from `git ls-files`, so the gitignored,
	// generated `.svelte-kit/` -- which is what actually resolves the
	// generated tsconfig that `extends` resolves -- never makes it in.
	// That was latent as long as every *.test.ts happened to only use
	// relative imports; the dry run's full `bun test` sweep now also picks
	// up src/routes/layout.test.ts (#251), which imports +layout.ts, which
	// imports #lib/settingsSync.js, and that fails to resolve with no
	// .svelte-kit/tsconfig.json in the sandbox. Stryker runs this directly
	// (no shell), so a `[ -f ... ] ||` guard isn't available -- `svelte-kit
	// sync` is itself already fast and idempotent (it just regenerates the
	// same routing manifest/types), so running it unconditionally per
	// mutant is the simpler correct choice.
	buildCommand: 'bunx svelte-kit sync',
	// Scoped to the modules the unit-test layer (#228) actually covers --
	// mutating a .svelte component or a route file Stryker can't run
	// bun:test against would just report every mutant as uncovered noise.
	mutate: [
		'src/lib/format.ts',
		'src/lib/flakyRuns.ts',
		'src/lib/relativeTime.ts',
		'src/lib/pipelinesFilters.svelte.ts',
		'src/lib/forgeFilter.svelte.ts',
		'src/lib/savedForgeTokens.ts',
		'src/lib/lastForgeSelection.ts',
	],
	// A surviving mutant fails the run (rules/javascript.md): 100, not a floor
	// to ratchet up later.
	thresholds: { high: 100, low: 100, break: 100 },
	concurrency: 4,
	reporters: ['clear-text', 'html'],
};
