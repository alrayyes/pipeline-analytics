module.exports = {
	ci: {
		collect: {
			url: ['http://localhost:4191/login'],
			numberOfRuns: 3,
			settings: {
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
