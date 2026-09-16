<script lang="ts">
import { defaultChartPadding, LineChart } from 'layerchart';
import { onMount } from 'svelte';
import { page } from '$app/state';
import { Badge } from '$lib/components/ui/badge/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';

interface Trend {
	timestamps: string[];
	p50?: number[];
	p90?: number[];
	rate?: number[];
}

interface PipelineDetail {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
	durationTrend: Trend;
	failureRateTrend: Trend;
}

const SIGNAL_LABELS: Record<string, string> = {
	failure_rate: 'elevated failure rate',
	duration_regression: 'duration regression',
	flaky_step: 'flaky step',
};

let detail = $state<PipelineDetail | null>(null);
let notFound = $state(false);
let error = $state<string | null>(null);

async function load(): Promise<void> {
	try {
		const res = await fetch(`/api/pipelines/${page.params.id}`);
		if (res.status === 404) {
			notFound = true;

			return;
		}

		if (!res.ok) {
			error = 'Could not load pipeline.';

			return;
		}

		detail = await res.json();
	} catch {
		error = 'Could not reach the server.';
	}
}

onMount(load);

function signalLabel(signal: string): string {
	return SIGNAL_LABELS[signal] ?? signal;
}

// LayerChart takes one row per point; the API returns parallel arrays, one
// value per bucket in each of timestamps/p50/p90/rate.
function durationSeries(trend: Trend) {
	return trend.timestamps.map((ts, i) => ({
		ts: new Date(ts),
		p50: trend.p50 ? trend.p50[i] / 60 : null,
		p90: trend.p90 ? trend.p90[i] / 60 : null,
	}));
}

function failureRateSeries(trend: Trend) {
	return trend.timestamps.map((ts, i) => ({
		ts: new Date(ts),
		rate: trend.rate ? trend.rate[i] * 100 : null,
	}));
}
</script>

<svelte:head>
	<title>{detail?.name ?? 'Pipeline'} · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-3xl px-4 py-8">
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
				variant={detail.healthStatus === 'healthy' ? 'default' : 'destructive'}
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

		<div class="mt-8 grid gap-6">
			<Card>
				<CardHeader>
					<CardTitle>Duration (p50 / p90, minutes)</CardTitle>
				</CardHeader>
				<CardContent>
					{#if detail.durationTrend.timestamps.length === 0}
						<p class="text-sm text-muted-foreground">Not enough run history yet.</p>
					{:else}
						<div class="h-[240px]">
							<LineChart
								data={durationSeries(detail.durationTrend)}
								x="ts"
								series={[
									{ key: 'p50', label: 'p50', color: 'var(--chart-1)' },
									{ key: 'p90', label: 'p90', color: 'var(--chart-2)' },
								]}
								padding={defaultChartPadding({ legend: true })}
								legend
							/>
						</div>
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
						<div class="h-[240px]">
							<LineChart
								data={failureRateSeries(detail.failureRateTrend)}
								x="ts"
								series={[{ key: 'rate', color: 'var(--chart-3)' }]}
							/>
						</div>
					{/if}
				</CardContent>
			</Card>
		</div>
	{/if}
</main>
