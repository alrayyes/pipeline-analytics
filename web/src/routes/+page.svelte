<script lang="ts">
import { onMount } from 'svelte';
import { Badge } from '$lib/components/ui/badge/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';

interface PipelineSummary {
	id: string;
	repoId: string;
	name: string;
	healthStatus: 'healthy' | 'unhealthy';
	triggeredSignals?: string[];
}

const SIGNAL_LABELS: Record<string, string> = {
	failure_rate: 'elevated failure rate',
	duration_regression: 'duration regression',
	flaky_step: 'flaky step',
};

let pipelines = $state<PipelineSummary[] | null>(null);
let error = $state<string | null>(null);

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

onMount(loadPipelines);

function signalLabel(signal: string): string {
	return SIGNAL_LABELS[signal] ?? signal;
}
</script>

<svelte:head>
	<title>pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="text-2xl font-semibold">Pipelines</h1>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if pipelines === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if pipelines.length === 0}
		<p class="mt-6 text-muted-foreground">
			No pipelines tracked yet. <a href="/repos" class="underline hover:text-foreground"
				>Register a repository</a
			> to start ingesting its runs.
		</p>
	{:else}
		<ul class="mt-6 grid gap-4">
			{#each pipelines as pipeline (pipeline.id)}
				<li>
					<a href="/pipelines/{pipeline.id}" class="block">
						<Card class="transition-colors hover:border-primary">
							<CardHeader class="flex flex-row items-center justify-between">
								<CardTitle class="contents">
									<h2>{pipeline.name}</h2>
								</CardTitle>
								<Badge
									variant={pipeline.healthStatus === 'healthy' ? 'default' : 'destructive'}
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
	{/if}
</main>
