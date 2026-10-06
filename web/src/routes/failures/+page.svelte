<script lang="ts">
import { getBranch } from '#lib/branchFilter.svelte.js';
import BranchSelect from '#lib/components/telemetry/BranchSelect.svelte';
import CategoryBreakdown from '#lib/components/telemetry/CategoryBreakdown.svelte';
import FailureGroupCard from '#lib/components/telemetry/FailureGroupCard.svelte';
import WindowToggle from '#lib/components/telemetry/WindowToggle.svelte';
import {
	type FailureInsights,
	fetchFailureInsights,
	type InsightsWindow,
} from '#lib/dashboardApi.js';
import {
	getTelemetryWindow,
	setTelemetryWindow,
} from '#lib/telemetryWindow.svelte.js';

let insights = $state<FailureInsights | null>(null);
let error = $state(false);
let loading = $state(true);

// Bumped on every window change and on teardown, so a slow response for a
// window the reader has left is dropped.
let generation = 0;

async function load(
	gen: number,
	window: InsightsWindow | undefined,
	branch: string | undefined,
): Promise<void> {
	try {
		const body = await fetchFailureInsights({ window, branch });
		if (gen !== generation) return;

		insights = body;
		error = false;
	} catch {
		if (gen !== generation) return;

		insights = null;
		error = true;
	}

	loading = false;
}

$effect(() => {
	const window = getTelemetryWindow() ?? undefined;
	const branch = getBranch() ?? undefined;
	const gen = ++generation;

	loading = true;
	load(gen, window, branch);

	return () => {
		generation++;
	};
});
</script>

<svelte:head>
	<title>Failures · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Root cause diagnostics</h1>
		<WindowToggle value={getTelemetryWindow() ?? insights?.window ?? null} onChange={setTelemetryWindow} />
		<BranchSelect window={getTelemetryWindow() ?? undefined} />
	</div>

	{#if loading}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error}
		<p role="alert" class="mt-6 text-destructive">Couldn't load failures right now.</p>
	{:else if insights && getBranch() && insights.totalRuns === 0}
		<p class="mt-6 text-muted-foreground">No runs on {getBranch()} in this window.</p>
	{:else if insights && insights.failureGroups.length === 0}
		<p class="mt-6 text-muted-foreground">No failures in this window.</p>
	{:else if insights}
		<CategoryBreakdown categories={insights.categoryBreakdown} />

		<h2 class="mt-8 text-lg font-semibold">Failing steps</h2>
		<ul class="mt-3 grid gap-4">
			{#each insights.failureGroups as group (`${group.step}|${group.category}|${group.conclusion ?? ''}`)}
				<li class="min-w-0">
					<FailureGroupCard {group} />
				</li>
			{/each}
		</ul>
	{/if}
</main>
