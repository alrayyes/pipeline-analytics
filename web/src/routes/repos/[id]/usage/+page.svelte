<script lang="ts">
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
import {
	ApiError,
	fetchRepoUsage,
	type UsageEntry,
} from '$lib/dashboardApi.js';

let usage = $state<UsageEntry[] | null>(null);
let notFound = $state(false);
let error = $state<string | null>(null);

async function load(): Promise<void> {
	try {
		usage = await fetchRepoUsage(page.params.id ?? '');
	} catch (e) {
		if (e instanceof ApiError && e.status === 404) {
			notFound = true;
		} else if (e instanceof ApiError) {
			error = 'Could not load usage.';
		} else {
			error = 'Could not reach the server.';
		}
	}
}

onMount(load);
</script>

<svelte:head>
	<title>Usage · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-3xl px-4 py-8">
	<a href="/" class="text-sm text-muted-foreground hover:underline">&larr; All pipelines</a>

	<h1 class="mt-4 text-2xl font-semibold">Runner-minutes usage</h1>

	<div class="mt-8">
		<Card>
			<CardHeader>
				<CardTitle>By workflow</CardTitle>
			</CardHeader>
			<CardContent>
				{#if error}
					<p role="alert" class="text-destructive">{error}</p>
				{:else if notFound}
					<p role="alert" class="text-destructive">Repository not found.</p>
				{:else if usage === null}
					<p class="text-sm text-muted-foreground">Loading…</p>
				{:else if usage.length === 0}
					<p class="text-sm text-muted-foreground">Not enough run history yet.</p>
				{:else}
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Workflow</TableHead>
								<TableHead>Runner minutes</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each usage as entry (entry.workflow)}
								<TableRow>
									<TableCell class="font-medium">{entry.workflow}</TableCell>
									<TableCell>{entry.runnerMinutes.toFixed(1)}</TableCell>
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				{/if}
			</CardContent>
		</Card>
	</div>
</main>
