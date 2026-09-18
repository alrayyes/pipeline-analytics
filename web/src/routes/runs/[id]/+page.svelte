<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import { onMount } from 'svelte';
import { page } from '$app/state';
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

interface RunStep {
	name: string;
	status: string;
	conclusion?: string;
	forgeUrl?: string;
}

interface RunDetail {
	runId: string;
	startedAt?: string;
	steps: RunStep[];
}

let detail = $state<RunDetail | null>(null);
let notFound = $state(false);
let error = $state<string | null>(null);

// Carried over from the flaky-runs list (#216) so this page's back link
// returns there instead of always falling back to the Pipelines overview.
const backHref = $derived.by(() => {
	const pipelineId = page.url.searchParams.get('pipelineId');
	const step = page.url.searchParams.get('step');

	return pipelineId && step
		? `/pipelines/${pipelineId}/flaky-runs?step=${encodeURIComponent(step)}`
		: '/';
});
const backLabel = $derived(
	page.url.searchParams.get('pipelineId')
		? 'Back to flaky runs'
		: 'All pipelines',
);

async function load(): Promise<void> {
	try {
		const res = await fetch(`/api/runs/${page.params.id}/steps`);
		if (res.status === 404) {
			notFound = true;

			return;
		}

		if (!res.ok) {
			error = 'Could not load run.';

			return;
		}

		detail = await res.json();
	} catch {
		error = 'Could not reach the server.';
	}
}

onMount(load);

function formatDate(iso?: string): string {
	return iso ? new Date(iso).toLocaleString() : 'Unknown run time';
}
</script>

<svelte:head>
	<title>Run · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-3xl px-4 py-8">
	<a href={backHref} class="text-sm text-muted-foreground hover:underline">
		&larr; {backLabel}
	</a>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if notFound}
		<p role="alert" class="mt-6 text-destructive">Run not found.</p>
	{:else if detail === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else}
		<h1 class="mt-4 text-2xl font-semibold">{formatDate(detail.startedAt)}</h1>

		<div class="mt-8">
			<Card>
				<CardHeader>
					<CardTitle>Steps</CardTitle>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Step</TableHead>
								<TableHead>Status</TableHead>
								<TableHead class="sr-only">Forge link</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each detail.steps as step, i (i)}
								<TableRow>
									<TableCell class="max-w-[16rem] truncate font-medium" title={step.name}>
										{step.name}
									</TableCell>
									<TableCell>{step.conclusion || step.status}</TableCell>
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
		</div>
	{/if}
</main>
