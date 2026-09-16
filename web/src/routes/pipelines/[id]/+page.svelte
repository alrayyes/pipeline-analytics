<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
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
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '$lib/components/ui/table/index.js';

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

interface Step {
	id: string;
	name: string;
	durationContributionSeconds: number;
	queueSeconds: number;
	execSeconds: number;
	failureRate: number;
	flaky: boolean;
	forgeUrl?: string;
}

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

async function loadSteps(): Promise<void> {
	try {
		const res = await fetch(`/api/pipelines/${page.params.id}/steps`);
		if (!res.ok) {
			stepsError = 'Could not load step breakdown.';

			return;
		}

		steps = await res.json();
	} catch {
		stepsError = 'Could not reach the server.';
	}
}

onMount(() => {
	load();
	loadSteps();
});

function signalLabel(signal: string): string {
	return SIGNAL_LABELS[signal] ?? signal;
}

function formatSeconds(seconds: number): string {
	return seconds >= 60
		? `${(seconds / 60).toFixed(1)}m`
		: `${seconds.toFixed(0)}s`;
}

function formatRate(rate: number): string {
	return `${Math.round(rate * 100)}%`;
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

			<Card>
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
									<TableRow>
										<TableCell class="font-medium">{step.name}</TableCell>
										<TableCell>{formatSeconds(step.durationContributionSeconds)}</TableCell>
										<TableCell>
											{formatSeconds(step.queueSeconds)} / {formatSeconds(step.execSeconds)}
										</TableCell>
										<TableCell>{formatRate(step.failureRate)}</TableCell>
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
											{/if}
										</TableCell>
										<TableCell>
											{#if step.forgeUrl}
												<a
													href={step.forgeUrl}
													target="_blank"
													rel="noreferrer"
													class="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground hover:underline"
												>
													<ExternalLinkIcon class="size-3.5" aria-hidden="true" />
													<span>View on forge</span>
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
