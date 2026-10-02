<script lang="ts">
import { onMount } from 'svelte';
import { page } from '$app/state';
import ForgeFilter from '$lib/components/ForgeFilter.svelte';
import { Badge } from '$lib/components/ui/badge/index.js';
import { Button } from '$lib/components/ui/button/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';
import { Label } from '$lib/components/ui/label/index.js';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
} from '$lib/components/ui/select/index.js';
import {
	ToggleGroup,
	ToggleGroupItem,
} from '$lib/components/ui/toggle-group/index.js';
import {
	ApiError,
	fetchPipelines,
	fetchRepos,
	type PipelineGroup,
	type PipelineSummary,
	type Repo,
} from '$lib/dashboardApi.js';
import { getForgeFilter } from '$lib/forgeFilter.svelte.js';
import {
	getHealthFilter,
	getRepoSelector,
	getSortBy,
	type HealthFilter,
	isAtDefaults,
	resetFilters,
	type SortBy,
	setHealthFilter,
	setRepoSelector,
	setSortBy,
} from '$lib/pipelinesFilters.svelte.js';
import { formatRelativeTime } from '$lib/relativeTime.js';

// Mirrors the Repos page's PAGE_SIZE (#165/#166).
const PAGE_SIZE = 20;

const SIGNAL_LABELS: Record<string, string> = {
	failure_rate: 'elevated failure rate',
	duration_regression: 'duration regression',
	flaky_step: 'flaky step',
};
const FORGE_LABELS: Record<string, string> = {
	github: 'GitHub',
	forgejo: 'Forgejo',
};
const HEALTH_OPTIONS: { value: HealthFilter; label: string }[] = [
	{ value: 'all', label: 'All' },
	{ value: 'healthy', label: 'Healthy' },
	{ value: 'unhealthy', label: 'Unhealthy' },
];
const SORT_LABELS: Record<SortBy, string> = {
	name: 'Name',
	lastRun: 'Most recently run',
};

let pipelines = $state<PipelineSummary[] | null>(null);
let repos = $state<Repo[] | null>(null);
let error = $state<string | null>(null);
let offset = $state(0);
let hasMore = $state(false);
// False only when a filtered load came back empty and an unfiltered probe
// found nothing either -- i.e. nothing has been ingested at all.
let anyPipelines = $state(true);
let loadSequence = 0;

// Grouped by repo (#102) -- a pipeline name alone ("CI") is ambiguous
// across more than one tracked repo, so each repo's pipelines get their
// own section rather than one flat list. Falls back to the bare repoId as
// the label if /api/repos hasn't loaded (or failed) -- grouping still
// works, it just can't show a human-readable name yet.
const repoById = $derived(new Map((repos ?? []).map((r) => [r.id, r])));

// Forge, repo, health filter and sort are all pushed down to the server
// (#243, #376): `pipelines` is one page of the full result, so anything
// applied after the fact would make a page's size and hasMore wrong.
const groupedPipelines = $derived.by(() => {
	const groups = new Map<string, PipelineGroup>();

	for (const pipeline of pipelines ?? []) {
		let group = groups.get(pipeline.repoId);

		if (!group) {
			const repo = repoById.get(pipeline.repoId);
			group = {
				repoId: pipeline.repoId,
				label: repo?.identifier ?? pipeline.repoId,
				forge: repo?.forge,
				pipelines: [],
			};
			groups.set(pipeline.repoId, group);
		}

		group.pipelines.push(pipeline);
	}

	// Map insertion order: groups and their pipelines follow the server's order.
	return [...groups.values()];
});

async function loadPipelines(): Promise<void> {
	try {
		const params: {
			limit: number;
			offset: number;
			forge?: 'github' | 'forgejo';
			repoId?: string;
			health?: 'healthy' | 'unhealthy';
			sort: SortBy;
		} = {
			limit: PAGE_SIZE,
			offset,
			sort: getSortBy(),
		};

		const filter = getForgeFilter();
		if (filter !== 'all') params.forge = filter;

		const selectedRepoId = getRepoSelector();
		if (selectedRepoId !== 'all') params.repoId = selectedRepoId;

		const health = getHealthFilter();
		if (health !== 'all') params.health = health;

		const sequence = ++loadSequence;
		const filtered =
			health !== 'all' || filter !== 'all' || selectedRepoId !== 'all';
		const body = await fetchPipelines(params);
		const empty = body.pipelines.length === 0 && offset === 0;
		let any = !empty;

		// An empty result under a filter is ambiguous: filtered to nothing, or
		// nothing ingested. Ask the unfiltered question before rendering either.
		if (empty && filtered) {
			try {
				any =
					(await fetchPipelines({ limit: 1, offset: 0 })).pipelines.length > 0;
			} catch {
				any = false;
			}
		}

		if (sequence !== loadSequence) return;

		anyPipelines = any;
		pipelines = body.pipelines;
		hasMore = body.hasMore;
	} catch (e) {
		error =
			e instanceof ApiError
				? 'Could not load pipelines.'
				: 'Could not reach the server.';
	}
}

async function loadRepos(): Promise<void> {
	repos = await fetchRepos();
}

function goToPreviousPage(): void {
	offset = Math.max(0, offset - PAGE_SIZE);
}

function goToNextPage(): void {
	if (!hasMore) return;

	offset += PAGE_SIZE;
}

// Tracks the forge filter's, repo selector's, health filter's and sort's
// previous values across effect runs so a real change to any (not just re-running for some
// other reason) is the only thing that resets the page -- changing either
// with the reader on page 2+ shouldn't leave them on an offset the
// newly-filtered result set might not even reach. Kept as its own effect,
// separate from the one that actually fetches below, same reasoning as
// the Repos page's identically-shaped `previousFilter` effect.
let previousForgeFilter: ReturnType<typeof getForgeFilter> | undefined;
let previousRepoSelector: string | undefined;
let previousHealthFilter: HealthFilter | undefined;
let previousSortBy: SortBy | undefined;

$effect(() => {
	const filter = getForgeFilter();
	const selectedRepoId = getRepoSelector();
	const health = getHealthFilter();
	const sort = getSortBy();

	if (
		filter !== previousForgeFilter ||
		selectedRepoId !== previousRepoSelector ||
		health !== previousHealthFilter ||
		sort !== previousSortBy
	) {
		previousForgeFilter = filter;
		previousRepoSelector = selectedRepoId;
		previousHealthFilter = health;
		previousSortBy = sort;
		offset = 0;
	}
});

// Reload with the current filters, sort, and page applied
// server-side -- replaces the earlier onMount(loadPipelines), and also
// covers the initial load.
$effect(() => {
	getForgeFilter();
	getRepoSelector();
	getHealthFilter();
	getSortBy();
	offset;
	loadPipelines();
});

onMount(() => {
	loadRepos();
});

function signalLabel(signal: string): string {
	return SIGNAL_LABELS[signal] ?? signal;
}
</script>

<svelte:head>
	<title>pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Pipelines</h1>
		<ForgeFilter />
	</div>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if pipelines === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if pipelines.length === 0 && offset === 0 && !anyPipelines}
		<p class="mt-6 text-muted-foreground">
			{#if page.data.hasRepos}
				No pipeline runs ingested yet. New runs arrive by webhook, plus a
				periodic check for anything missed -- this can take a few minutes
				after registering a repository.
			{:else}
				No repositories registered yet -- use "Register a repository" above
				to start tracking one.
			{/if}
		</p>
	{:else}
		<div class="mt-6 flex flex-wrap items-center gap-4">
			<div role="radiogroup" aria-label="Filter by health status">
				<ToggleGroup
					type="single"
					variant="outline"
					value={getHealthFilter()}
					onValueChange={(value) => {
						if (value) setHealthFilter(value as HealthFilter);
					}}
				>
					{#each HEALTH_OPTIONS as option (option.value)}
						<ToggleGroupItem value={option.value} aria-label={option.label}>
							{option.label}
						</ToggleGroupItem>
					{/each}
				</ToggleGroup>
			</div>

			<div class="flex items-center gap-2">
				<Label for="repo-filter">Repo</Label>
				<Select
					type="single"
					value={getRepoSelector()}
					onValueChange={(value) => {
						if (value) setRepoSelector(value);
					}}
				>
					<SelectTrigger id="repo-filter" class="w-48">
						{getRepoSelector() === 'all'
							? 'All repos'
							: (repoById.get(getRepoSelector())?.identifier ?? getRepoSelector())}
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all" label="All repos">All repos</SelectItem>
						{#each repos ?? [] as repo (repo.id)}
							<SelectItem value={repo.id} label={repo.identifier}>
								{repo.identifier}
							</SelectItem>
						{/each}
					</SelectContent>
				</Select>
			</div>

			<div class="flex items-center gap-2">
				<Label for="sort-by">Sort</Label>
				<Select
					type="single"
					value={getSortBy()}
					onValueChange={(value) => {
						if (value) setSortBy(value as SortBy);
					}}
				>
					<SelectTrigger id="sort-by" class="w-44">
						{SORT_LABELS[getSortBy()]}
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="name" label={SORT_LABELS.name}>{SORT_LABELS.name}</SelectItem>
						<SelectItem value="lastRun" label={SORT_LABELS.lastRun}>
							{SORT_LABELS.lastRun}
						</SelectItem>
					</SelectContent>
				</Select>
			</div>

			<Button
				variant="outline"
				size="sm"
				disabled={isAtDefaults()}
				onclick={resetFilters}
			>
				Reset filters
			</Button>
		</div>

		{#if pipelines.length === 0}
			<p class="mt-6 text-muted-foreground">No pipelines match the selected filters.</p>
		{:else}
			<p class="mt-4 text-sm text-muted-foreground">
				Showing {pipelines.length}
				{pipelines.length === 1 ? 'pipeline' : 'pipelines'} on this page.
			</p>

			<div class="mt-4 grid gap-8">
				{#each groupedPipelines as group (group.repoId)}
					<section class="min-w-0">
						<div class="mb-3 flex items-center gap-2">
							<h2 class="min-w-0 truncate text-sm font-semibold text-muted-foreground">
								{group.label}
							</h2>
							{#if group.forge}
								<Badge variant="outline">{FORGE_LABELS[group.forge] ?? group.forge}</Badge>
							{/if}
						</div>
						<ul class="grid gap-4">
							{#each group.pipelines as pipeline (pipeline.id)}
								<li class="min-w-0">
									<a href="/pipelines/{pipeline.id}" class="block">
										<Card class="min-w-0 transition-colors hover:border-primary">
											<CardHeader class="flex flex-row items-center justify-between">
												<CardTitle class="contents">
													<h3 class="min-w-0 truncate">{pipeline.name}</h3>
												</CardTitle>
												<Badge
													variant={pipeline.healthStatus === 'healthy'
														? 'success'
														: 'destructive'}
													class={pipeline.healthStatus === 'unhealthy'
														? 'bg-destructive text-white'
														: ''}
												>
													{pipeline.healthStatus}
												</Badge>
											</CardHeader>
											{#if pipeline.triggeredSignals?.length || pipeline.lastRunAt}
												<CardContent class="grid gap-1 text-sm text-muted-foreground">
													{#if pipeline.triggeredSignals?.length}
														<p>{pipeline.triggeredSignals.map(signalLabel).join(', ')}</p>
													{/if}
													{#if pipeline.lastRunAt}
														<p class="text-xs">
															Last run {formatRelativeTime(new Date(pipeline.lastRunAt))}
														</p>
													{/if}
												</CardContent>
											{/if}
										</Card>
									</a>
								</li>
							{/each}
						</ul>
					</section>
				{/each}
			</div>

			<div class="mt-6 flex items-center justify-between gap-4">
				<Button
					variant="outline"
					size="sm"
					disabled={offset === 0}
					onclick={goToPreviousPage}
				>
					Previous
				</Button>
				<Button variant="outline" size="sm" disabled={!hasMore} onclick={goToNextPage}>
					Next
				</Button>
			</div>
		{/if}
	{/if}
</main>
