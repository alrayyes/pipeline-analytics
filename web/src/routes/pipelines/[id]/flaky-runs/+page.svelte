<script lang="ts">
import { onMount } from 'svelte';
import { goto } from '$app/navigation';
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

interface FlakyRun {
	runId: string;
	startedAt?: string;
}

let runs = $state<FlakyRun[] | null>(null);
let notFound = $state(false);
let error = $state<string | null>(null);

const stepName = $derived(page.url.searchParams.get('step') ?? '');
const pipelineId = $derived(page.params.id ?? '');

async function load(step: string): Promise<void> {
	runs = null;

	try {
		const res = await fetch(
			`/api/pipelines/${page.params.id}/flaky-runs?step=${encodeURIComponent(step)}`,
		);
		if (res.status === 404) {
			notFound = true;

			return;
		}

		if (!res.ok) {
			error = 'Could not load flaky runs.';

			return;
		}

		runs = await res.json();
	} catch {
		error = 'Could not reach the server.';
	}
}

onMount(() => {
	load(stepName);
});

// runHref carries the pipeline/step context along so the run page's own
// back link can return here instead of falling all the way back to the
// Pipelines overview.
function runHref(run: FlakyRun): string {
	const params = new URLSearchParams({ pipelineId, step: stepName });

	return `/runs/${run.runId}?${params}`;
}

function formatDate(iso?: string): string {
	return iso ? new Date(iso).toLocaleString() : 'Unknown';
}
</script>

<svelte:head>
	<title>Flaky runs · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-3xl px-4 py-8">
	<a
		href="/pipelines/{page.params.id}"
		class="text-sm text-muted-foreground hover:underline"
	>
		&larr; Back to pipeline
	</a>

	<h1 class="mt-4 text-2xl font-semibold">Flaky runs</h1>
	<p class="mt-1 truncate text-sm text-muted-foreground" title={stepName}>{stepName}</p>

	<div class="mt-8">
		<Card>
			<CardHeader>
				<CardTitle>Runs where this step failed</CardTitle>
			</CardHeader>
			<CardContent>
				{#if error}
					<p role="alert" class="text-destructive">{error}</p>
				{:else if notFound}
					<p role="alert" class="text-destructive">Pipeline not found.</p>
				{:else if runs === null}
					<p class="text-sm text-muted-foreground">Loading…</p>
				{:else if runs.length === 0}
					<p class="text-sm text-muted-foreground">No flaky runs in this window.</p>
				{:else}
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Run</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each runs as run (run.runId)}
								<TableRow class="cursor-pointer" onclick={() => goto(runHref(run))}>
									<TableCell>
										<a href={runHref(run)} class="hover:underline">
											{formatDate(run.startedAt)}
										</a>
									</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				{/if}
			</CardContent>
		</Card>
	</div>
</main>
