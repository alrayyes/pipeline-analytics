// Shared fetch layer for the dashboard's read views -- pulled out of the
// pages themselves (#296) so a WebMCP tool and the page that renders the
// same data call one function instead of the tool growing its own,
// second path to the same API.

export interface PipelineSummary {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
	lastRunAt?: string;
}

export interface Repo {
	id: string;
	forge: 'github' | 'forgejo';
	identifier: string;
}

export interface PipelineGroup {
	repoId: string;
	label: string;
	forge?: 'github' | 'forgejo';
	pipelines: PipelineSummary[];
}

export interface PipelineListResponse {
	pipelines: PipelineSummary[];
	hasMore: boolean;
}

export interface Trend {
	timestamps: string[];
	p50?: number[];
	p90?: number[];
	rate?: number[];
}

export interface PipelineDetail {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
	durationTrend: Trend;
	failureRateTrend: Trend;
}

export interface Step {
	id: string;
	name: string;
	durationContributionSeconds: number;
	queueSeconds: number;
	execSeconds: number;
	failureRate: number;
	failureCount: number;
	flaky: boolean;
	forgeUrl?: string;
}

// What a run's or step's forge state means, decided by the server
// (metrics.OutcomeOf). The raw status and conclusion ride along but a client
// shouldn't interpret them.
export type Outcome =
	| 'passed'
	| 'failed'
	| 'running'
	| 'queued'
	| 'cancelled'
	| 'skipped'
	| 'unknown';

export interface RunStep {
	name: string;
	status: string;
	conclusion?: string;
	outcome: Outcome;
	forgeUrl?: string;
}

export type RunStatusFilter = 'all' | 'failed' | 'running' | 'success';
export type InsightsWindow = '24h' | '7d' | '30d';

export type FailureCategory =
	| 'infrastructure'
	| 'code_tests'
	| 'network_timeouts'
	| 'config_secrets'
	| 'uncategorised';

export interface RunSummary {
	id: string;
	pipelineId: string;
	pipelineName: string;
	repoId: string;
	status: string;
	conclusion?: string;
	outcome: Outcome;
	startedAt?: string;
	durationSeconds?: number;
	// Commit fields are absent on runs ingested before they were recorded.
	branch?: string;
	sha?: string;
	message?: string;
	actor?: string;
	forgeUrl?: string;
	steps: RunStep[];
}

export interface RunListResponse {
	runs: RunSummary[];
	hasMore: boolean;
}

export interface StageFailureCount {
	step: string;
	failures: number;
	// Fraction of all failed-step occurrences in the window, from the server.
	share: number;
}

export interface CategoryCount {
	category: FailureCategory;
	occurrences: number;
	// Fraction of all failed-step occurrences in the window, from the server.
	share: number;
}

export interface FailingPipeline {
	pipelineId: string;
	pipelineName: string;
	repoId: string;
	runs: number;
	failedRuns: number;
}

export interface FailureGroup {
	step: string;
	category: FailureCategory;
	conclusion?: string;
	occurrences: number;
	pipelines: { pipelineId: string; pipelineName: string }[];
}

export interface FailureInsights {
	// The window these figures cover: the one requested, or the server's own
	// default when none was. Show this; don't assume a default.
	window: InsightsWindow;
	totalRuns: number;
	failedRuns: number;
	// Absent, not zero, when no run concluded in the window.
	passRate?: number;
	// Percentage points against the preceding window; absent without one.
	passRateDelta?: number;
	flakyStepRatio: number;
	// Absent when nothing recovered in the window.
	mttrSeconds?: number;
	stageDistribution: StageFailureCount[];
	categoryBreakdown: CategoryCount[];
	topFailingPipelines: FailingPipeline[];
	failureGroups: FailureGroup[];
}

export interface FlakyStep {
	pipelineId: string;
	pipelineName: string;
	repoId: string;
	name: string;
	// Fraction of the step's runs in the window that failed, from the server.
	flakeRate: number;
	// Every run of the step in the window, not just the ones in the matrix.
	runCount: number;
	// The step's most recent results (at most 40), oldest first.
	recentOutcomes: Outcome[];
}

export interface FlakyStepList {
	// The window these figures cover; show this, don't assume a default.
	window: InsightsWindow;
	steps: FlakyStep[];
	hasMore: boolean;
}

export interface UsageEntry {
	workflow: string;
	runnerMinutes: number;
}

// Carries the response status a caller needs to tell "not found" apart from
// any other failure (the pipeline and usage pages both do), without every
// caller re-parsing a generic Error's message to get it back.
export class ApiError extends Error {
	status: number;

	constructor(status: number) {
		super(`Request failed with status ${status}`);
		this.status = status;
	}
}

export interface PipelinesParams {
	limit: number;
	offset: number;
	forge?: 'github' | 'forgejo';
	repoId?: string;
	// Both are applied by the server across every matching pipeline before
	// paging; leave them off for the default (every status, by name).
	health?: 'healthy' | 'unhealthy';
	sort?: 'name' | 'lastRun';
}

export async function fetchPipelines(
	params: PipelinesParams,
	fetchFn: typeof fetch = fetch,
): Promise<PipelineListResponse> {
	const searchParams = new URLSearchParams({
		limit: String(params.limit),
		offset: String(params.offset),
	});
	if (params.forge) searchParams.set('forge', params.forge);
	if (params.repoId) searchParams.set('repoId', params.repoId);
	if (params.health) searchParams.set('health', params.health);
	if (params.sort) searchParams.set('sort', params.sort);

	const res = await fetchFn(`/api/pipelines?${searchParams}`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as PipelineListResponse;
}

// Best effort, matching the overview page's existing behavior: repo names
// are a grouping nicety, not something worth surfacing an error banner for.
export async function fetchRepos(
	fetchFn: typeof fetch = fetch,
): Promise<Repo[]> {
	try {
		const res = await fetchFn('/api/repos');
		if (!res.ok) return [];

		const body = (await res.json()) as { repos: Repo[] };
		return body.repos;
	} catch {
		return [];
	}
}

export async function fetchPipeline(
	id: string,
	fetchFn: typeof fetch = fetch,
): Promise<PipelineDetail> {
	const res = await fetchFn(`/api/pipelines/${id}`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as PipelineDetail;
}

export async function fetchPipelineSteps(
	id: string,
	fetchFn: typeof fetch = fetch,
): Promise<Step[]> {
	const res = await fetchFn(`/api/pipelines/${id}/steps`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as Step[];
}

export async function fetchRepoUsage(
	repoId: string,
	fetchFn: typeof fetch = fetch,
): Promise<UsageEntry[]> {
	const res = await fetchFn(`/api/repos/${repoId}/usage`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as UsageEntry[];
}

export interface FailureInsightsParams {
	// Omit to let the server use its default window.
	window?: InsightsWindow;
	forge?: 'github' | 'forgejo';
	repoId?: string;
}

export async function fetchFailureInsights(
	params: FailureInsightsParams,
	fetchFn: typeof fetch = fetch,
): Promise<FailureInsights> {
	const searchParams = new URLSearchParams();
	if (params.window) searchParams.set('window', params.window);
	if (params.repoId) searchParams.set('repoId', params.repoId);
	if (params.forge) searchParams.set('forge', params.forge);

	const query = searchParams.size > 0 ? `?${searchParams}` : '';
	const res = await fetchFn(`/api/insights/failures${query}`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as FailureInsights;
}

export interface RunsParams {
	limit: number;
	offset: number;
	status?: RunStatusFilter;
	forge?: 'github' | 'forgejo';
	repoId?: string;
}

export async function fetchRuns(
	params: RunsParams,
	fetchFn: typeof fetch = fetch,
): Promise<RunListResponse> {
	const searchParams = new URLSearchParams({
		limit: String(params.limit),
		offset: String(params.offset),
	});
	// "all" is the server's default, so it's left off the URL.
	if (params.status && params.status !== 'all') {
		searchParams.set('status', params.status);
	}
	if (params.repoId) searchParams.set('repoId', params.repoId);
	if (params.forge) searchParams.set('forge', params.forge);

	const res = await fetchFn(`/api/runs?${searchParams}`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as RunListResponse;
}

export interface FlakyStepsParams {
	// Omit to let the server use its default window.
	window?: InsightsWindow;
	forge?: 'github' | 'forgejo';
	repoId?: string;
	limit?: number;
	offset?: number;
}

export async function fetchFlakySteps(
	params: FlakyStepsParams,
	fetchFn: typeof fetch = fetch,
): Promise<FlakyStepList> {
	const searchParams = new URLSearchParams();
	if (params.window) searchParams.set('window', params.window);
	if (params.repoId) searchParams.set('repoId', params.repoId);
	if (params.forge) searchParams.set('forge', params.forge);
	if (params.limit !== undefined) {
		searchParams.set('limit', String(params.limit));
	}
	if (params.offset !== undefined) {
		searchParams.set('offset', String(params.offset));
	}

	const query = searchParams.size > 0 ? `?${searchParams}` : '';
	const res = await fetchFn(`/api/steps/flaky${query}`);
	if (!res.ok) throw new ApiError(res.status);

	return (await res.json()) as FlakyStepList;
}
