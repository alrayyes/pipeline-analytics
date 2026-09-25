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
