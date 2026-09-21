<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import HistoryIcon from '@lucide/svelte/icons/history';
import { onMount } from 'svelte';
import ForgeFilter from '$lib/components/ForgeFilter.svelte';
import { Badge } from '$lib/components/ui/badge/index.js';
import { Button } from '$lib/components/ui/button/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '$lib/components/ui/table/index.js';
import { flakyRunsHref } from '$lib/flakyRuns.js';
import { getForgeFilter } from '$lib/forgeFilter.svelte.js';
import { formatRate, formatSeconds } from '$lib/format.js';

interface Step {
	id: string;
	name: string;
	durationContributionSeconds: number;
	failureRate: number;
	failureCount: number;
	flaky: boolean;
	forgeUrl?: string;
}

interface PipelineStepsGroup {
	pipelineId: string;
	pipelineName: string;
	repoId: string;
	steps: Step[];
}

interface Repo {
	id: string;
	forge: 'github' | 'forgejo';
	identifier: string;
}

interface RepoGroup {
	repoId: string;
	label: string;
	forge?: 'github' | 'forgejo';
	pipelines: PipelineStepsGroup[];
}

interface UnhealthyStepsResponse {
	groups: PipelineStepsGroup[];
	hasMore: boolean;
}

// Mirrors the Pipelines page's PAGE_SIZE (#250).
const PAGE_SIZE = 20;

const FORGE_LABELS: Record<string, string> = {
	github: 'GitHub',
	forgejo: 'Forgejo',
};

let groups = $state<PipelineStepsGroup[] | null>(null);
let repos = $state<Repo[] | null>(null);
let error = $state<string | null>(null);
let offset = $state(0);
let hasMore = $state(false);

const repoById = $derived(new Map((repos ?? []).map((r) => [r.id, r])));

// The forge filter stays client-side, unlike the Pipelines page's own
// (#250) -- there's no repos.forge join to push it into here, since
// pagination happens after per-pipeline step aggregation in Go, not in
// SQL (design.md's "Paginate the computed result, not the candidate
// fetch"). Applied to the current page's groups only, so a filter change
// can show fewer results than a full page, or none, even when a matching
// pipeline exists on another page.
const forgeFilteredGroups = $derived.by(() => {
	const filter = getForgeFilter();
	if (filter === 'all' || !groups) return groups;

	return groups.filter((g) => repoById.get(g.repoId)?.forge === filter);
});

// Grouped by repo, same reasoning as the Pipelines list (#102): a pipeline
// name alone is ambiguous across more than one tracked repo.
const groupedByRepo = $derived.by(() => {
	const byRepo = new Map<string, RepoGroup>();

	for (const pipeline of forgeFilteredGroups ?? []) {
		let group = byRepo.get(pipeline.repoId);

		if (!group) {
			const repo = repoById.get(pipeline.repoId);
			group = {
				repoId: pipeline.repoId,
				label: repo?.identifier ?? pipeline.repoId,
				forge: repo?.forge,
				pipelines: [],
			};
			byRepo.set(pipeline.repoId, group);
		}

		group.pipelines.push(pipeline);
	}

	return [...byRepo.values()].sort((a, b) => a.label.localeCompare(b.label));
});

async function loadUnhealthySteps(): Promise<void> {
	try {
		const params = new URLSearchParams({
			limit: String(PAGE_SIZE),
			offset: String(offset),
		});

		const res = await fetch(`/api/steps/unhealthy?${params}`);
		if (!res.ok) {
			error = 'Could not load unhealthy steps.';

			return;
		}

		const body: UnhealthyStepsResponse = await res.json();
		groups = body.groups;
		hasMore = body.hasMore;
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

function goToPreviousPage(): void {
	offset = Math.max(0, offset - PAGE_SIZE);
}

function goToNextPage(): void {
	if (!hasMore) return;

	offset += PAGE_SIZE;
}

// A real change to the forge filter (not just re-running for some other
// reason) resets the page -- changing it with the reader on page 2+
// shouldn't leave them on an offset the newly-filtered result set might
// not even reach. Same reasoning as the Pipelines page's identically
// shaped effect.
let previousForgeFilter: ReturnType<typeof getForgeFilter> | undefined;

$effect(() => {
	const filter = getForgeFilter();

	if (filter !== previousForgeFilter) {
		previousForgeFilter = filter;
		offset = 0;
	}
});

$effect(() => {
	offset;
	loadUnhealthySteps();
});

onMount(() => {
	loadRepos();
});
</script>

<svelte:head>
	<title>Steps · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Unhealthy steps</h1>
		<ForgeFilter />
	</div>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if groups === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if forgeFilteredGroups?.length === 0}
		<p class="mt-6 text-muted-foreground">
			{#if groups.length === 0 && offset === 0}
				No flaky or failing steps right now.
			{:else if groups.length === 0}
				No flaky or failing steps on this page.
			{:else}
				No unhealthy steps match the selected forge filter.
			{/if}
		</p>
	{:else}
		<div class="mt-6 grid gap-8">
			{#each groupedByRepo as group (group.repoId)}
				<section class="min-w-0">
					<div class="mb-3 flex items-center gap-2">
						<h2 class="min-w-0 truncate text-sm font-semibold text-muted-foreground">
							{group.label}
						</h2>
						{#if group.forge}
							<Badge variant="outline">{FORGE_LABELS[group.forge] ?? group.forge}</Badge>
						{/if}
					</div>
					<div class="grid gap-4">
						{#each group.pipelines as pipeline (pipeline.pipelineId)}
							<Card class="min-w-0">
								<CardHeader>
									<CardTitle class="contents">
										<h3 class="min-w-0 truncate">
											<a href="/pipelines/{pipeline.pipelineId}" class="hover:underline">
												{pipeline.pipelineName}
											</a>
										</h3>
									</CardTitle>
								</CardHeader>
								<CardContent>
									<Table>
										<TableHeader>
											<TableRow>
												<TableHead>Step</TableHead>
												<TableHead>Duration</TableHead>
												<TableHead>Failure rate</TableHead>
												<TableHead>Status</TableHead>
												<TableHead class="sr-only">Forge link</TableHead>
											</TableRow>
										</TableHeader>
										<TableBody>
											{#each pipeline.steps as step (step.id)}
												<TableRow>
													<TableCell class="max-w-[16rem] truncate font-medium" title={step.name}>
														{step.name}
													</TableCell>
													<TableCell>{formatSeconds(step.durationContributionSeconds)}</TableCell>
													<TableCell>{formatRate(step.failureRate)} ({step.failureCount})</TableCell>
													<TableCell>
														{#if step.flaky}
															<Badge
																variant="outline"
																class="border-amber-600 text-amber-700 dark:border-amber-400 dark:text-amber-400"
															>
																flaky
															</Badge>
														{:else}
															<Badge variant="destructive" class="bg-destructive text-white">
																failing
															</Badge>
														{/if}
													</TableCell>
													<TableCell>
														{#if step.flaky}
															<a
																href={flakyRunsHref(pipeline.pipelineId, step.name)}
																aria-label="View flaky runs"
																title="View flaky runs"
																class="inline-flex items-center text-muted-foreground hover:text-foreground"
															>
																<HistoryIcon class="size-4" aria-hidden="true" />
															</a>
														{:else if step.forgeUrl}
															<a
																href={step.forgeUrl}
																target="_blank"
																rel="noreferrer"
																aria-label="View on forge"
																title="View on forge"
																class="inline-flex items-center text-muted-foreground hover:text-foreground"
															>
																<ExternalLinkIcon class="size-4" aria-hidden="true" />
															</a>
														{/if}
													</TableCell>
												</TableRow>
											{/each}
										</TableBody>
									</Table>
								</CardContent>
							</Card>
						{/each}
					</div>
				</section>
			{/each}
		</div>

		<div class="mt-6 flex items-center justify-between gap-4">
			<Button variant="outline" size="sm" disabled={offset === 0} onclick={goToPreviousPage}>
				Previous
			</Button>
			<Button variant="outline" size="sm" disabled={!hasMore} onclick={goToNextPage}>
				Next
			</Button>
		</div>
	{/if}
</main>
