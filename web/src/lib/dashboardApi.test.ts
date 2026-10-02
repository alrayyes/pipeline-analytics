import { describe, expect, test } from 'bun:test';
import {
	ApiError,
	fetchFailureInsights,
	fetchPipeline,
	fetchPipelineSteps,
	fetchPipelines,
	fetchRepos,
	fetchRepoUsage,
	fetchRuns,
	type PipelineDetail,
} from './dashboardApi.js';

type FetchMock = (input: RequestInfo | URL) => Promise<Response>;

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), { status });
}

describe('fetchPipelines', () => {
	test('calls GET /api/pipelines with the given filter and pagination params', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(
				jsonResponse({
					pipelines: [
						{ id: '1', repoId: 'r1', name: 'CI', healthStatus: 'healthy' },
					],
					hasMore: true,
				}),
			);
		};

		const result = await fetchPipelines(
			{ limit: 20, offset: 20, forge: 'github', repoId: 'r1' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe(
			'/api/pipelines?limit=20&offset=20&forge=github&repoId=r1',
		);
		expect(result).toEqual({
			pipelines: [
				{ id: '1', repoId: 'r1', name: 'CI', healthStatus: 'healthy' },
			],
			hasMore: true,
		});
	});

	test('omits forge and repoId when not given', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse({ pipelines: [], hasMore: false }));
		};

		await fetchPipelines({ limit: 20, offset: 0 }, fetchFn as typeof fetch);

		expect(requestedUrl).toBe('/api/pipelines?limit=20&offset=0');
	});

	test('throws an ApiError on a non-OK response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 500 }));

		await expect(
			fetchPipelines({ limit: 20, offset: 0 }, fetchFn as typeof fetch),
		).rejects.toThrow(ApiError);
	});
});

describe('fetchRepos', () => {
	test('returns the parsed repo list', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(
				jsonResponse({
					repos: [{ id: '1', forge: 'github', identifier: 'org/repo' }],
				}),
			);

		await expect(fetchRepos(fetchFn as typeof fetch)).resolves.toEqual([
			{ id: '1', forge: 'github', identifier: 'org/repo' },
		]);
	});

	test('resolves to an empty list on a non-OK response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 500 }));

		await expect(fetchRepos(fetchFn as typeof fetch)).resolves.toEqual([]);
	});

	test('resolves to an empty list when the request itself fails', async () => {
		const fetchFn: FetchMock = () => Promise.reject(new Error('network error'));

		await expect(fetchRepos(fetchFn as typeof fetch)).resolves.toEqual([]);
	});
});

describe('fetchPipeline', () => {
	test('calls GET /api/pipelines/{id} and returns the parsed detail', async () => {
		let requestedUrl: string | undefined;
		const detail: PipelineDetail = {
			id: '1',
			repoId: 'r1',
			name: 'CI',
			healthStatus: 'healthy',
			durationTrend: { timestamps: [] },
			failureRateTrend: { timestamps: [] },
		};
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse(detail));
		};

		await expect(fetchPipeline('1', fetchFn as typeof fetch)).resolves.toEqual(
			detail,
		);
		expect(requestedUrl).toBe('/api/pipelines/1');
	});

	test('throws an ApiError carrying the response status on a non-OK response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 404 }));

		try {
			await fetchPipeline('missing', fetchFn as typeof fetch);
			throw new Error('expected fetchPipeline to throw');
		} catch (e) {
			if (!(e instanceof ApiError)) throw e;
			expect(e.status).toBe(404);
		}
	});
});

describe('fetchPipelineSteps', () => {
	test('calls GET /api/pipelines/{id}/steps and returns the parsed step list', async () => {
		let requestedUrl: string | undefined;
		const steps = [
			{
				id: 's1',
				name: 'build',
				durationContributionSeconds: 10,
				queueSeconds: 1,
				execSeconds: 9,
				failureRate: 0,
				failureCount: 0,
				flaky: false,
			},
		];
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse(steps));
		};

		await expect(
			fetchPipelineSteps('1', fetchFn as typeof fetch),
		).resolves.toEqual(steps);
		expect(requestedUrl).toBe('/api/pipelines/1/steps');
	});

	test('throws an ApiError on a non-OK response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 500 }));

		await expect(
			fetchPipelineSteps('1', fetchFn as typeof fetch),
		).rejects.toThrow(ApiError);
	});
});

describe('fetchRepoUsage', () => {
	test('calls GET /api/repos/{repoId}/usage and returns the parsed usage list', async () => {
		let requestedUrl: string | undefined;
		const usage = [{ workflow: 'CI', runnerMinutes: 12.5 }];
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse(usage));
		};

		await expect(
			fetchRepoUsage('r1', fetchFn as typeof fetch),
		).resolves.toEqual(usage);
		expect(requestedUrl).toBe('/api/repos/r1/usage');
	});

	test('throws an ApiError carrying the response status on a non-OK response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(new Response(null, { status: 404 }));

		try {
			await fetchRepoUsage('missing', fetchFn as typeof fetch);
			throw new Error('expected fetchRepoUsage to throw');
		} catch (e) {
			if (!(e instanceof ApiError)) throw e;
			expect(e.status).toBe(404);
		}
	});
});

describe('fetchFailureInsights', () => {
	const insights = {
		totalRuns: 10,
		failedRuns: 2,
		flakyStepRatio: 0.1,
		stageDistribution: [],
		categoryBreakdown: [],
		topFailingPipelines: [],
		failureGroups: [],
	};

	test('calls GET /api/insights/failures with the window', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse(insights));
		};

		const result = await fetchFailureInsights(
			{ window: '30d' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe('/api/insights/failures?window=30d');
		expect(result).toEqual(insights);
	});

	test('adds repoId and forge only when given', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse(insights));
		};

		await fetchFailureInsights(
			{ window: '24h', repoId: 'r1', forge: 'forgejo' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe(
			'/api/insights/failures?window=24h&repoId=r1&forge=forgejo',
		);
	});

	test('a missing passRate stays absent rather than becoming zero', async () => {
		const fetchFn: FetchMock = () => Promise.resolve(jsonResponse(insights));

		const result = await fetchFailureInsights(
			{ window: '7d' },
			fetchFn as typeof fetch,
		);

		expect(result.passRate).toBeUndefined();
	});

	test('throws an ApiError carrying the status on a non-2xx response', async () => {
		const fetchFn: FetchMock = () =>
			Promise.resolve(jsonResponse({ message: 'nope' }, 401));

		const error = await fetchFailureInsights(
			{ window: '7d' },
			fetchFn as typeof fetch,
		).catch((e: unknown) => e);

		expect(error).toBeInstanceOf(ApiError);
		expect((error as ApiError).status).toBe(401);
	});
});

describe('fetchRuns', () => {
	test('calls GET /api/runs with the status and pagination params', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse({ runs: [], hasMore: false }));
		};

		await fetchRuns(
			{ limit: 20, offset: 40, status: 'failed' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe('/api/runs?limit=20&offset=40&status=failed');
	});

	test('omits status when it is "all", the server default', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse({ runs: [], hasMore: false }));
		};

		await fetchRuns(
			{ limit: 20, offset: 0, status: 'all' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe('/api/runs?limit=20&offset=0');
	});

	test('adds repoId, forge and pipeline scope only when given', async () => {
		let requestedUrl: string | undefined;
		const fetchFn: FetchMock = (input) => {
			requestedUrl = String(input);
			return Promise.resolve(jsonResponse({ runs: [], hasMore: false }));
		};

		await fetchRuns(
			{ limit: 10, offset: 0, repoId: 'r1', forge: 'github' },
			fetchFn as typeof fetch,
		);

		expect(requestedUrl).toBe(
			'/api/runs?limit=10&offset=0&repoId=r1&forge=github',
		);
	});

	test('returns runs with their steps and optional commit fields', async () => {
		const run = {
			id: 'run-1',
			pipelineId: 'p1',
			pipelineName: 'CI',
			repoId: 'r1',
			status: 'completed',
			conclusion: 'failure',
			sha: 'c4d291a',
			steps: [{ name: 'build', status: 'completed', conclusion: 'success' }],
		};
		const fetchFn: FetchMock = () =>
			Promise.resolve(jsonResponse({ runs: [run], hasMore: true }));

		const result = await fetchRuns(
			{ limit: 1, offset: 0 },
			fetchFn as typeof fetch,
		);

		expect(result.runs[0]).toEqual(run);
		expect(result.runs[0].branch).toBeUndefined();
		expect(result.hasMore).toBe(true);
	});

	test('throws an ApiError carrying the status on a non-2xx response', async () => {
		const fetchFn: FetchMock = () => Promise.resolve(jsonResponse({}, 400));

		const error = await fetchRuns(
			{ limit: 1, offset: 0 },
			fetchFn as typeof fetch,
		).catch((e: unknown) => e);

		expect(error).toBeInstanceOf(ApiError);
		expect((error as ApiError).status).toBe(400);
	});
});
