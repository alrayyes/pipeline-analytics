<script lang="ts">
import { onMount } from 'svelte';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '#lib/components/ui/card/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '#lib/components/ui/table/index.js';
import {
	ApiError,
	fetchGitHubRateLimit,
	type GitHubTokenUsage,
	type RateLimitStatus,
} from '#lib/dashboardApi.js';
import { formatRelativeTime } from '#lib/relativeTime.js';

let tokens = $state<GitHubTokenUsage[] | null>(null);
let error = $state<string | null>(null);

async function load(): Promise<void> {
	try {
		tokens = await fetchGitHubRateLimit();
	} catch (e) {
		error =
			e instanceof ApiError
				? 'Could not load GitHub API usage.'
				: 'Could not reach the server.';
	}
}

onMount(load);

function usagePercent(status: RateLimitStatus): number {
	return status.limit === 0
		? 0
		: Math.round((status.used / status.limit) * 100);
}
</script>

<svelte:head>
	<title>Insights · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="text-2xl font-semibold">Insights</h1>

	<div class="mt-8">
		<Card>
			<CardHeader>
				<CardTitle>GitHub API rate limit</CardTitle>
			</CardHeader>
			<CardContent>
				{#if error}
					<p role="alert" class="text-destructive">{error}</p>
				{:else if tokens === null}
					<p class="text-sm text-muted-foreground">Loading…</p>
				{:else if tokens.length === 0}
					<p class="text-sm text-muted-foreground">
						No GitHub repositories tracked yet.
					</p>
				{:else}
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Token</TableHead>
								<TableHead>Repositories</TableHead>
								<TableHead>Used / limit</TableHead>
								<TableHead>Remaining</TableHead>
								<TableHead>Refreshes</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{#each tokens as token (token.tokenMasked + token.repos.join(','))}
								<TableRow>
									<TableCell class="font-medium">{token.tokenMasked}</TableCell>
									<TableCell class="max-w-[16rem] truncate" title={token.repos.join(', ')}>
										{token.repos.join(', ')}
									</TableCell>
									{#if token.status}
										<TableCell>
											{token.status.used} / {token.status.limit} ({usagePercent(
												token.status,
											)}%)
										</TableCell>
										<TableCell>{token.status.remaining}</TableCell>
										<TableCell>
											{formatRelativeTime(new Date(token.status.resetAt))}
										</TableCell>
									{:else}
										<TableCell colspan={3} class="text-muted-foreground">
											Not observed yet -- refreshes after the next reconciliation poll.
										</TableCell>
									{/if}
								</TableRow>
							{/each}
						</TableBody>
					</Table>
				{/if}
			</CardContent>
		</Card>
	</div>
</main>
