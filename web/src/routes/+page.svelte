<script lang="ts">
import { onMount } from 'svelte';
import { page } from '$app/state';
import ForgeFilter from '$lib/components/ForgeFilter.svelte';
import { Badge } from '$lib/components/ui/badge/index.js';
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
import { getForgeFilter } from '$lib/forgeFilter.svelte.js';
import { formatRelativeTime } from '$lib/relativeTime.js';

interface PipelineSummary {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
	lastRunAt?: string;
}

interface Repo {
	id: string;
	forge: 'github' | 'forgejo';
	identifier: string;
}

interface PipelineGroup {
	repoId: string;
	label: string;
	forge?: 'github' | 'forgejo';
	pipelines: PipelineSummary[];
}

type HealthFilter = 'all' | 'healthy' | 'unhealthy';
type SortBy = 'name' | 'lastRun';

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

// Defaults to unhealthy-only (#101): this is a monitoring dashboard, so
// leading with what needs attention beats an everything-at-once list --
// per dashboard filtering research, surfacing only what needs attention
// first reads better on load than mixing it into everything that's fine.
let healthFilter = $state<HealthFilter>('unhealthy');
let selectedRepoId = $state('all');
let sortBy = $state<SortBy>('name');

// Grouped by repo (#102) -- a pipeline name alone ("CI") is ambiguous
// across more than one tracked repo, so each repo's pipelines get their
// own section rather than one flat list. Falls back to the bare repoId as
// the label if /api/repos hasn't loaded (or failed) -- grouping still
// works, it just can't show a human-readable name yet.
const repoById = $derived(new Map((repos ?? []).map((r) => [r.id, r])));

// The forge filter is shared with the Repos page (issue #140), but applied
// client-side here rather than as a query param -- unlike Repos, this page
// already does a full unpaginated fetch and joins forge in from repoById,
// so there's no server round trip to save by pushing the filter down.
const forgeFilteredPipelines = $derived.by(() => {
	const filter = getForgeFilter();
	if (filter === 'all' || !pipelines) return pipelines;

	return pipelines.filter((p) => repoById.get(p.repoId)?.forge === filter);
});

const repoFilteredPipelines = $derived.by(() => {
	if (selectedRepoId === 'all' || !forgeFilteredPipelines)
		return forgeFilteredPipelines;

	return forgeFilteredPipelines.filter((p) => p.repoId === selectedRepoId);
});

const visiblePipelines = $derived.by(() => {
	if (!repoFilteredPipelines) return null;
	if (healthFilter === 'all') return repoFilteredPipelines;

	return repoFilteredPipelines.filter((p) => p.healthStatus === healthFilter);
});

// "Name" is an explicit alphabetical sort rather than whatever order
// /api/pipelines happened to return -- the point of offering it as a choice
// is that it's deterministic, the same way "Most recently run" is.
function comparePipelines(a: PipelineSummary, b: PipelineSummary): number {
	if (sortBy === 'lastRun') {
		if (!a.lastRunAt && !b.lastRunAt) return 0;
		if (!a.lastRunAt) return 1;
		if (!b.lastRunAt) return -1;

		return new Date(b.lastRunAt).getTime() - new Date(a.lastRunAt).getTime();
	}

	return a.name.localeCompare(b.name);
}

const groupedPipelines = $derived.by(() => {
	const groups = new Map<string, PipelineGroup>();

	for (const pipeline of visiblePipelines ?? []) {
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

	for (const group of groups.values()) {
		group.pipelines.sort(comparePipelines);
	}

	return [...groups.values()].sort((a, b) => a.label.localeCompare(b.label));
});

async function loadPipelines(): Promise<void> {
	try {
		const res = await fetch('/api/pipelines');
		if (!res.ok) {
			error = 'Could not load pipelines.';
			return;
		}

		pipelines = await res.json();
	} catch {
		error = 'Could not reach the server.';
	}
}

async function loadRepos(): Promise<void> {
	try {
		const res = await fetch('/api/repos');
		if (!res.ok) return;

		const body: { repos: Repo[] } = await res.json();
		repos = body.repos;
	} catch {
		// Best effort -- grouping still works, just with repoId as the label.
	}
}

onMount(() => {
	loadPipelines();
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
	{:else if pipelines.length === 0}
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
					value={healthFilter}
					onValueChange={(value) => {
						if (value) healthFilter = value as HealthFilter;
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
					value={selectedRepoId}
					onValueChange={(value) => {
						if (value) selectedRepoId = value;
					}}
				>
					<SelectTrigger id="repo-filter" class="w-48">
						{selectedRepoId === 'all'
							? 'All repos'
							: (repoById.get(selectedRepoId)?.identifier ?? selectedRepoId)}
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
					value={sortBy}
					onValueChange={(value) => {
						if (value) sortBy = value as SortBy;
					}}
				>
					<SelectTrigger id="sort-by" class="w-44">
						{SORT_LABELS[sortBy]}
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="name" label={SORT_LABELS.name}>{SORT_LABELS.name}</SelectItem>
						<SelectItem value="lastRun" label={SORT_LABELS.lastRun}>
							{SORT_LABELS.lastRun}
						</SelectItem>
					</SelectContent>
				</Select>
			</div>
		</div>

		{#if visiblePipelines?.length === 0}
			<p class="mt-6 text-muted-foreground">No pipelines match the selected filters.</p>
		{:else}
			<p class="mt-4 text-sm text-muted-foreground">
				{#if healthFilter === 'all'}
					Showing all {visiblePipelines?.length ?? 0}
					{visiblePipelines?.length === 1 ? 'pipeline' : 'pipelines'}.
				{:else}
					Showing {visiblePipelines?.length ?? 0} {healthFilter} of {repoFilteredPipelines?.length ??
						0}.
				{/if}
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
		{/if}
	{/if}
</main>
