<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import HistoryIcon from '@lucide/svelte/icons/history';
import { defaultChartPadding, LineChart } from 'layerchart';
import { onMount } from 'svelte';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import { Badge } from '$lib/components/ui/badge/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';
import {
	type ChartConfig,
	ChartContainer,
	ChartTooltip,
} from '$lib/components/ui/chart/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '$lib/components/ui/table/index.js';
import {
	ApiError,
	fetchPipeline,
	fetchPipelineSteps,
	type PipelineDetail,
	type Step,
	type Trend,
} from '$lib/dashboardApi.js';
import { flakyRunsHref } from '$lib/flakyRuns.js';
import { formatRate, formatSeconds } from '$lib/format.js';

// Categorical slots 1/2/8 of the validated default data-viz palette
// (light/dark both fully specified, adjacent pairs pre-validated for CVD
// separation) -- the previous --chart-1/--chart-2 CSS variables are the
// shadcn scaffold's own placeholder values (pure grayscale, zero chroma),
// never actually replaced.
const DURATION_CHART_CONFIG: ChartConfig = {
	p50: { label: 'p50', theme: { light: '#2a78d6', dark: '#3987e5' } },
	p90: { label: 'p90', theme: { light: '#eb6834', dark: '#d95926' } },
};
const FAILURE_RATE_CHART_CONFIG: ChartConfig = {
	rate: { label: 'Failure rate', theme: { light: '#e34948', dark: '#e66767' } },
};

const SIGNAL_LABELS: Record<string, string> = {
	failure_rate: 'elevated failure rate',
	duration_regression: 'duration regression',
	flaky_step: 'flaky step',
};

let detail = $state<PipelineDetail | null>(null);
let notFound = $state(false);
let error = $state<string | null>(null);
let steps = $state<Step[] | null>(null);
let stepsError = $state<string | null>(null);

const pipelineId = $derived(page.params.id ?? '');

async function load(): Promise<void> {
	try {
		detail = await fetchPipeline(pipelineId);
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) {
			notFound = true;
		} else if (e instanceof ApiError) {
			error = 'Could not load pipeline.';
		} else {
			error = 'Could not reach the server.';
		}
	}
}

async function loadSteps(): Promise<void> {
	try {
		steps = await fetchPipelineSteps(pipelineId);
	} catch (e) {
		stepsError =
			e instanceof ApiError
				? 'Could not load step breakdown.'
				: 'Could not reach the server.';
	}
}

onMount(() => {
	load();
	loadSteps();
});

function signalLabel(signal: string): string {
	return SIGNAL_LABELS[signal] ?? signal;
}

// LayerChart takes one row per point; the API returns parallel arrays, one
// value per bucket in each of timestamps/p50/p90/rate.
// Rounded to a sane display precision -- otherwise the tooltip shows a raw
// floating-point minute value (5.083333333333333) instead of something
// readable.
function durationSeries(trend: Trend) {
	return trend.timestamps.map((ts, i) => ({
		ts: new Date(ts),
		p50: trend.p50 ? Math.round((trend.p50[i] / 60) * 10) / 10 : null,
		p90: trend.p90 ? Math.round((trend.p90[i] / 60) * 10) / 10 : null,
	}));
}

function failureRateSeries(trend: Trend) {
	return trend.timestamps.map((ts, i) => ({
		ts: new Date(ts),
		rate: trend.rate ? Math.round(trend.rate[i] * 1000) / 10 : null,
	}));
}
</script>

<svelte:head>
	<title>{detail?.name ?? 'Pipeline'} · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<a href="/" class="text-sm text-muted-foreground hover:underline">&larr; All pipelines</a>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if notFound}
		<p role="alert" class="mt-6 text-destructive">Pipeline not found.</p>
	{:else if detail === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else}
		<div class="mt-4 flex items-center justify-between">
			<h1 class="text-2xl font-semibold">{detail.name}</h1>
			<Badge
				variant={detail.healthStatus === 'healthy' ? 'success' : 'destructive'}
				class={detail.healthStatus === 'unhealthy' ? 'bg-destructive text-white' : ''}
			>
				{detail.healthStatus}
			</Badge>
		</div>
		{#if detail.triggeredSignals?.length}
			<p class="mt-1 text-sm text-muted-foreground">
				{detail.triggeredSignals.map(signalLabel).join(', ')}
			</p>
		{/if}
		<a
			href="/repos/{detail.repoId}/usage"
			class="mt-1 inline-block text-sm text-muted-foreground hover:underline"
		>
			Runner-minutes usage &rarr;
		</a>

		<div class="mt-8 grid gap-6">
			<Card>
				<CardHeader>
					<CardTitle>Duration (p50 / p90, minutes)</CardTitle>
				</CardHeader>
				<CardContent>
					{#if detail.durationTrend.timestamps.length === 0}
						<p class="text-sm text-muted-foreground">Not enough run history yet.</p>
					{:else}
						<ChartContainer config={DURATION_CHART_CONFIG} class="aspect-auto h-[240px] w-full">
							<LineChart
								data={durationSeries(detail.durationTrend)}
								x="ts"
								series={[
									{ key: 'p50', label: 'p50', color: 'var(--color-p50)' },
									{ key: 'p90', label: 'p90', color: 'var(--color-p90)' },
								]}
								padding={defaultChartPadding({ legend: true })}
								legend
							>
								{#snippet tooltip()}
									<ChartTooltip labelFormatter={(value) => value.toLocaleDateString()} />
								{/snippet}
							</LineChart>
						</ChartContainer>
					{/if}
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>Failure rate (%)</CardTitle>
				</CardHeader>
				<CardContent>
					{#if detail.failureRateTrend.timestamps.length === 0}
						<p class="text-sm text-muted-foreground">Not enough run history yet.</p>
					{:else}
						<ChartContainer
							config={FAILURE_RATE_CHART_CONFIG}
							class="aspect-auto h-[240px] w-full"
						>
							<LineChart
								data={failureRateSeries(detail.failureRateTrend)}
								x="ts"
								series={[{ key: 'rate', color: 'var(--color-rate)' }]}
							>
								{#snippet tooltip()}
									<ChartTooltip labelFormatter={(value) => value.toLocaleDateString()} />
								{/snippet}
							</LineChart>
						</ChartContainer>
					{/if}
				</CardContent>
			</Card>

			<Card class="min-w-0">
				<CardHeader>
					<CardTitle>Steps</CardTitle>
				</CardHeader>
				<CardContent>
					{#if stepsError}
						<p role="alert" class="text-destructive">{stepsError}</p>
					{:else if steps === null}
						<p class="text-sm text-muted-foreground">Loading…</p>
					{:else if steps.length === 0}
						<p class="text-sm text-muted-foreground">Not enough run history yet.</p>
					{:else}
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Step</TableHead>
									<TableHead>Duration</TableHead>
									<TableHead>Queue / exec</TableHead>
									<TableHead>Failure rate</TableHead>
									<TableHead>Status</TableHead>
									<TableHead class="sr-only">Forge link</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each steps as step (step.id)}
									<TableRow
										class={step.flaky || step.forgeUrl ? 'cursor-pointer' : ''}
										onclick={step.flaky
											? () => goto(flakyRunsHref(pipelineId, step.name))
											: step.forgeUrl
												? (event: MouseEvent) => {
														// Skip when the click already came from the real
														// link -- it already navigated, so opening a
														// second tab/page here would be a duplicate.
														if ((event.target as HTMLElement).closest('a')) return;

														window.open(step.forgeUrl, '_blank', 'noopener,noreferrer');
													}
												: undefined}
									>
										<TableCell class="max-w-[16rem] truncate font-medium" title={step.name}>
											{step.name}
										</TableCell>
										<TableCell>{formatSeconds(step.durationContributionSeconds)}</TableCell>
										<TableCell>
											{formatSeconds(step.queueSeconds)} / {formatSeconds(step.execSeconds)}
										</TableCell>
										<TableCell>{formatRate(step.failureRate)} ({step.failureCount})</TableCell>
										<TableCell>
											{#if step.flaky}
												<Badge
													variant="outline"
													class="border-amber-600 text-amber-700 dark:border-amber-400 dark:text-amber-400"
												>
													flaky
												</Badge>
											{:else if step.failureRate > 0}
												<Badge variant="destructive" class="bg-destructive text-white">
													failing
												</Badge>
											{:else}
												<Badge variant="success">passing</Badge>
											{/if}
										</TableCell>
										<TableCell>
											{#if step.flaky}
												<a
													href={flakyRunsHref(pipelineId, step.name)}
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
					{/if}
				</CardContent>
			</Card>
		</div>
	{/if}
</main>
