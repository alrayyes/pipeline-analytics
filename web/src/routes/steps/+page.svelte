<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import { onMount } from 'svelte';
import ForgeFilter from '$lib/components/ForgeFilter.svelte';
import { Badge } from '$lib/components/ui/badge/index.js';
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
import { getForgeFilter } from '$lib/forgeFilter.svelte.js';

interface Step {
	id: string;
	name: string;
	durationContributionSeconds: number;
	failureRate: number;
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

const FORGE_LABELS: Record<string, string> = {
	github: 'GitHub',
	forgejo: 'Forgejo',
};

let groups = $state<PipelineStepsGroup[] | null>(null);
let repos = $state<Repo[] | null>(null);
let error = $state<string | null>(null);

const repoById = $derived(new Map((repos ?? []).map((r) => [r.id, r])));

// Same client-side forge filter as the Pipelines list (#140), shared across
// both pages -- this page's fetch is unpaginated too, so there's no server
// round trip to save by pushing it down.
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
		const res = await fetch('/api/steps/unhealthy');
		if (!res.ok) {
			error = 'Could not load unhealthy steps.';

			return;
		}

		groups = await res.json();
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
	loadUnhealthySteps();
	loadRepos();
});

function formatSeconds(seconds: number): string {
	return seconds >= 60
		? `${(seconds / 60).toFixed(1)}m`
		: `${seconds.toFixed(0)}s`;
}

function formatRate(rate: number): string {
	return `${Math.round(rate * 100)}%`;
}
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
			{#if groups.length === 0}
				No flaky or failing steps right now.
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
													<TableCell>{formatRate(step.failureRate)}</TableCell>
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
														{#if step.forgeUrl}
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
	{/if}
</main>
