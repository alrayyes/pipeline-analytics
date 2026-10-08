module.exports = {
	ci: {
		collect: {
			// /login first, signed out; the rest need the session that
			// scripts/lighthouse-auth.cjs sets up (#520). One run each: the
			// 30 a three-run pass would take is too long for CI, and the
			// gated categories are deterministic.
			url: [
				'http://localhost:4191/login',
				'http://localhost:4191/',
				'http://localhost:4191/pipelines',
				'http://localhost:4191/failures',
				'http://localhost:4191/flaky',
				'http://localhost:4191/runs',
				'http://localhost:4191/insights',
				'http://localhost:4191/repos',
				'http://localhost:4191/settings',
				'http://localhost:4191/releases',
				'http://localhost:4191/legal',
			],
			numberOfRuns: 1,
			puppeteerScript: './scripts/lighthouse-auth.cjs',
			puppeteerLaunchOptions: { args: ['--no-sandbox'] },
			settings: {
				// Lighthouse clears cookies between runs by default, which
				// would sign the browser out again.
				disableStorageReset: true,
				// /login deliberately probes POST /api/auth/login/options and
				// treats a 404 as "no account registered yet, show the
				// registration flow" (internal/httpserver/auth.go's
				// ErrNoUser handling, web/src/routes/login/+page.svelte's
				// detectMode()) -- correct, already-handled application
				// behavior, but Chrome logs the underlying failed network
				// request to the console regardless of whether JS handles
				// it, and errors-in-console has no way to tell the two
				// apart. Skipped rather than thresholded around.
				skipAudits: ['errors-in-console'],
			},
		},
		assert: {
			assertions: {
				'categories:accessibility': ['error', { minScore: 0.95 }],
				'categories:best-practices': ['error', { minScore: 0.95 }],
				'categories:seo': ['error', { minScore: 0.9 }],
				// Performance is warn-only: shared CI runners make strict
				// performance gating flaky, same reasoning as this repo's
				// mutation-testing CI job reporting a score without gating
				// on one.
				'categories:performance': ['warn', { minScore: 0.8 }],
			},
		},
		upload: {
			target: 'filesystem',
			outputDir: './.lighthouseci',
		},
	},
};
