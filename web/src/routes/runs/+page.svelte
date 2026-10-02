<script lang="ts">
import RunCard from '$lib/components/telemetry/RunCard.svelte';
import { Button } from '$lib/components/ui/button/index.js';
import {
	ToggleGroup,
	ToggleGroupItem,
} from '$lib/components/ui/toggle-group/index.js';
import {
	fetchRuns,
	type RunStatusFilter,
	type RunSummary,
} from '$lib/dashboardApi.js';
import {
	getOffset,
	getStatus,
	nextPage,
	previousPage,
	RUNS_PAGE_SIZE,
	setStatus,
} from '$lib/runsFilters.svelte.js';
import { statusTone } from '$lib/statusModel.js';

const REFRESH_MS = 15_000;

const OPTIONS: { value: RunStatusFilter; label: string }[] = [
	{ value: 'all', label: 'All' },
	{ value: 'failed', label: 'Failed' },
	{ value: 'running', label: 'Running' },
	{ value: 'success', label: 'Success' },
];

let runs = $state<RunSummary[] | null>(null);
let hasMore = $state(false);
let error = $state(false);
let loading = $state(true);

let timer: ReturnType<typeof setTimeout> | undefined;
// Bumped on every filter/page change and on teardown, so a slow response (or
// a refresh already scheduled) for a list the reader has left is dropped.
let generation = 0;

async function load(
	gen: number,
	params: { offset: number; status: RunStatusFilter },
	background: boolean,
): Promise<void> {
	try {
		const body = await fetchRuns({
			limit: RUNS_PAGE_SIZE,
			offset: params.offset,
			status: params.status === 'all' ? undefined : params.status,
		});
		if (gen !== generation) return;

		runs = body.runs;
		hasMore = body.hasMore;
		error = false;
	} catch {
		if (gen !== generation) return;

		// A failed background refresh keeps the list the reader is looking at.
		if (!background) {
			runs = null;
			error = true;
		}
	}

	if (gen !== generation) return;

	loading = false;

	if (runs?.some((r) => statusTone(r.status, r.conclusion) === 'running')) {
		timer = setTimeout(() => load(gen, params, true), REFRESH_MS);
	}
}

$effect(() => {
	const params = { offset: getOffset(), status: getStatus() };
	const gen = ++generation;

	loading = true;
	load(gen, params, false);

	return () => {
		generation++;
		clearTimeout(timer);
	};
});
</script>

<svelte:head>
	<title>Runs · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Runs</h1>
		<div role="radiogroup" aria-label="Filter by status">
			<ToggleGroup
				type="single"
				variant="outline"
				value={getStatus()}
				onValueChange={(value) => {
					if (value) setStatus(value as RunStatusFilter);
				}}
			>
				{#each OPTIONS as option (option.value)}
					<ToggleGroupItem value={option.value} aria-label={option.label}>
						{option.label}
					</ToggleGroupItem>
				{/each}
			</ToggleGroup>
		</div>
	</div>

	{#if loading}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error}
		<p role="alert" class="mt-6 text-destructive">Couldn't load runs right now.</p>
	{:else if runs?.length === 0}
		<p class="mt-6 text-muted-foreground">No runs match this filter.</p>
	{:else}
		<ul class="mt-6 grid gap-4">
			{#each runs ?? [] as run (run.id)}
				<li class="min-w-0">
					<RunCard {run} />
				</li>
			{/each}
		</ul>
	{/if}

	<div class="mt-6 flex items-center justify-between gap-4">
		<Button
			variant="outline"
			size="sm"
			disabled={getOffset() === 0}
			onclick={previousPage}
		>
			Previous
		</Button>
		<Button variant="outline" size="sm" disabled={!hasMore} onclick={nextPage}>
			Next
		</Button>
	</div>
</main>
