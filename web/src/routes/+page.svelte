<script lang="ts">
import ArrowDownIcon from '@lucide/svelte/icons/arrow-down';
import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
import { getBranch } from '#lib/branchFilter.svelte.js';
import BranchSelect from '#lib/components/telemetry/BranchSelect.svelte';
import MetricCard from '#lib/components/telemetry/MetricCard.svelte';
import WindowToggle from '#lib/components/telemetry/WindowToggle.svelte';
import {
	type FailureInsights,
	fetchFailureInsights,
	fetchRuns,
	type InsightsWindow,
	type RunSummary,
} from '#lib/dashboardApi.js';
import {
	formatPercentagePoints,
	formatRate,
	formatSeconds,
} from '#lib/format.js';
import {
	getTelemetryWindow,
	setTelemetryWindow,
} from '#lib/telemetryWindow.svelte.js';

let { data }: { data: { hasRepos: boolean } } = $props();

let insights = $state<FailureInsights | null>(null);
let latestFailure = $state<RunSummary | null>(null);
let error = $state(false);
let loading = $state(true);

// Bumped on every window change and on teardown, so a slow response for a
// window the reader has left is dropped.
let generation = 0;

function plural(count: number, one: string, many: string): string {
	return count === 1 ? one : many;
}

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
	if (!data.hasRepos) return;

	const window = getTelemetryWindow() ?? undefined;
	const branch = getBranch() ?? undefined;
	const gen = ++generation;

	loading = true;
	load(gen, window, branch);

	return () => {
		generation++;
	};
});

// The latest failure isn't windowed, but it follows the chosen branch. Failing
// to load it must not take the rest of the page down, so it's fetched on its
// own.
$effect(() => {
	if (!data.hasRepos) return;

	const branch = getBranch() ?? undefined;
	let stale = false;

	fetchRuns({ limit: 1, offset: 0, status: 'failed', branch })
		.then((body) => {
			if (!stale) latestFailure = body.runs[0] ?? null;
		})
		.catch(() => {
			if (!stale) latestFailure = null;
		});

	return () => {
		stale = true;
	};
});
</script>

<svelte:head>
	<title>Overview · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex flex-wrap items-center gap-4">
		<h1 class="text-2xl font-semibold">Overview</h1>
		{#if data.hasRepos}
			<WindowToggle value={getTelemetryWindow() ?? insights?.window ?? null} onChange={setTelemetryWindow} />
			<BranchSelect window={getTelemetryWindow() ?? undefined} />
		{/if}
	</div>

	{#if !data.hasRepos}
		<p class="mt-6 text-muted-foreground">
			No repositories registered yet -- use "Register a repository" above
			to start tracking one.
		</p>
	{:else}
		{#if latestFailure}
			<section
				aria-labelledby="latest-failure-heading"
				class="mt-6 flex min-w-0 flex-wrap items-center justify-between gap-3 rounded-lg border p-4"
			>
				<div class="grid min-w-0 gap-1">
					<h2 id="latest-failure-heading" class="text-xs text-muted-foreground">
						Latest failure
					</h2>
					<p class="flex min-w-0 flex-wrap items-center gap-2">
						<span class="font-medium break-words">{latestFailure.pipelineName}</span>
						{#if latestFailure.sha}
							<code class="text-sm">{latestFailure.sha.slice(0, 7)}</code>
						{/if}
					</p>
					{#if latestFailure.message}
						<p class="truncate text-sm text-muted-foreground">{latestFailure.message}</p>
					{/if}
				</div>
				<a href={`/runs/${latestFailure.id}`} class="text-sm font-medium underline">
					Inspect run
				</a>
			</section>
		{/if}

		{#if loading}
			<p class="mt-6 text-muted-foreground">Loading…</p>
		{:else if error}
			<p role="alert" class="mt-6 text-destructive">Couldn't load the overview right now.</p>
		{:else if insights && getBranch() && insights.totalRuns === 0}
			<p class="mt-6 text-muted-foreground">
				No runs on {getBranch()} in this window.
			</p>
		{:else if insights}
			<div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
				<MetricCard
					label="Pass rate"
					value={insights.passRate === undefined ? 'No data' : formatRate(insights.passRate)}
					note={insights.passRateDelta === undefined
						? undefined
						: formatPercentagePoints(insights.passRateDelta)}
					icon={insights.passRateDelta === undefined || insights.passRateDelta === 0
						? undefined
						: insights.passRateDelta < 0
							? ArrowDownIcon
							: ArrowUpIcon}
				/>
				<MetricCard
					label="Mean time to recovery"
					value={insights.mttrSeconds === undefined
						? 'No recoveries'
						: formatSeconds(insights.mttrSeconds)}
				/>
				<MetricCard label="Failed runs" value={`${insights.failedRuns} / ${insights.totalRuns}`} />
				<MetricCard label="Flaky steps" value={formatRate(insights.flakyStepRatio)} />
			</div>

			{#if insights.stageDistribution.length > 0}
				<section aria-labelledby="stage-heading" class="mt-8">
					<h2 id="stage-heading" class="text-lg font-semibold">Failure stage distribution</h2>
					<ul class="mt-3 grid gap-3">
						{#each insights.stageDistribution as stage (stage.step)}
							<li class="min-w-0">
								<div class="flex min-w-0 items-baseline gap-3 text-sm">
									<span class="min-w-0 flex-1 truncate">{stage.step}</span>
									<span class="shrink-0 text-muted-foreground">
										{stage.failures}
										{plural(stage.failures, 'failure', 'failures')}
									</span>
									<span class="shrink-0 font-medium">{formatRate(stage.share)}</span>
								</div>
								<div aria-hidden="true" class="mt-1 h-2 w-full overflow-hidden rounded-full bg-muted">
									<div class="h-full bg-foreground" style:width="{stage.share * 100}%"></div>
								</div>
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			{#if insights.topFailingPipelines.length > 0}
				<section aria-labelledby="top-heading" class="mt-8">
					<h2 id="top-heading" class="text-lg font-semibold">Top failing pipelines</h2>
					<ul class="mt-3 grid gap-2">
						{#each insights.topFailingPipelines as pipeline (pipeline.pipelineId)}
							<li class="flex min-w-0 flex-wrap items-baseline justify-between gap-x-3">
								<a href={`/pipelines/${pipeline.pipelineId}`} class="min-w-0 truncate font-medium underline">
									{pipeline.pipelineName}
								</a>
								<span class="text-sm text-muted-foreground">
									{pipeline.failedRuns} of {pipeline.runs}
									{plural(pipeline.runs, 'run', 'runs')} failed
								</span>
							</li>
						{/each}
					</ul>
				</section>
			{/if}
		{/if}
	{/if}
</main>
