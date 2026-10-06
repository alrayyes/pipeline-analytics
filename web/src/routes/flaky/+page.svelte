<script lang="ts">
import { getBranch } from '#lib/branchFilter.svelte.js';
import BranchSelect from '#lib/components/telemetry/BranchSelect.svelte';
import FlakyStepCard from '#lib/components/telemetry/FlakyStepCard.svelte';
import WindowToggle from '#lib/components/telemetry/WindowToggle.svelte';
import {
	type FlakyStepList,
	fetchFlakySteps,
	type InsightsWindow,
} from '#lib/dashboardApi.js';
import {
	getTelemetryWindow,
	setTelemetryWindow,
} from '#lib/telemetryWindow.svelte.js';

let list = $state<FlakyStepList | null>(null);
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
		const body = await fetchFlakySteps({ window, branch });
		if (gen !== generation) return;

		list = body;
		error = false;
	} catch {
		if (gen !== generation) return;

		list = null;
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
	<title>Flaky tests · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Flaky tests</h1>
		<WindowToggle value={getTelemetryWindow() ?? list?.window ?? null} onChange={setTelemetryWindow} />
		<BranchSelect window={getTelemetryWindow() ?? undefined} />
	</div>

	{#if loading}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error}
		<p role="alert" class="mt-6 text-destructive">Couldn't load flaky steps right now.</p>
	{:else if list && list.steps.length === 0}
		<p class="mt-6 text-muted-foreground">
			No flaky steps{getBranch() ? ` on ${getBranch()}` : ''} in this window.
		</p>
	{:else if list}
		<ul class="mt-6 grid gap-4">
			{#each list.steps as step (`${step.pipelineId}|${step.name}`)}
				<li class="min-w-0">
					<FlakyStepCard {step} />
				</li>
			{/each}
		</ul>
	{/if}
</main>
