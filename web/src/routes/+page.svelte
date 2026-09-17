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
import { getForgeFilter } from '$lib/forgeFilter.svelte.js';

interface PipelineSummary {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
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

const SIGNAL_LABELS: Record<string, string> = {
	failure_rate: 'elevated failure rate',
	duration_regression: 'duration regression',
	flaky_step: 'flaky step',
};
const FORGE_LABELS: Record<string, string> = {
	github: 'GitHub',
	forgejo: 'Forgejo',
};

let pipelines = $state<PipelineSummary[] | null>(null);
let repos = $state<Repo[] | null>(null);
let error = $state<string | null>(null);

// Defaults to unhealthy-only (#101): this is a monitoring dashboard, so
// leading with what needs attention beats an everything-at-once list --
// per dashboard filtering research, surfacing only what needs attention
// first reads better on load than mixing it into everything that's fine.
let showAll = $state(false);

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

const unhealthyPipelines = $derived(
	forgeFilteredPipelines?.filter((p) => p.healthStatus === 'unhealthy') ?? null,
);
const visiblePipelines = $derived(
	showAll ? forgeFilteredPipelines : unhealthyPipelines,
);

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

		repos = await res.json();
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
	{:else if forgeFilteredPipelines?.length === 0}
		<p class="mt-6 text-muted-foreground">
			No pipelines match the selected forge filter.
		</p>
	{:else}
		<div class="mt-6 flex items-center justify-between">
			<p class="text-sm text-muted-foreground">
				{#if showAll}
					Showing all {forgeFilteredPipelines?.length ?? 0}
					{forgeFilteredPipelines?.length === 1 ? 'pipeline' : 'pipelines'}.
				{:else}
					Showing {unhealthyPipelines?.length ?? 0} unhealthy of {forgeFilteredPipelines?.length ??
						0}.
				{/if}
			</p>
			<Button variant="outline" size="sm" onclick={() => (showAll = !showAll)}>
				{showAll ? 'Show unhealthy only' : 'Show all'}
			</Button>
		</div>

		{#if !showAll && unhealthyPipelines?.length === 0}
			<p class="mt-4 text-muted-foreground">
				All {forgeFilteredPipelines?.length ?? 0}
				{forgeFilteredPipelines?.length === 1 ? 'pipeline is' : 'pipelines are'} healthy.
			</p>
		{:else}
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
											{#if pipeline.triggeredSignals?.length}
												<CardContent class="text-sm text-muted-foreground">
													{pipeline.triggeredSignals.map(signalLabel).join(', ')}
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
