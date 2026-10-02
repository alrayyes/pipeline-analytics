<script lang="ts">
import DOMPurify from 'dompurify';
import { marked } from 'marked';
import { onMount } from 'svelte';
import type { Release } from '$lib/changelog.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';

// Generated from CHANGELOG.md at build time (scripts/generate-releases.ts)
// and served by the app itself, so opening this page never queries GitHub
// (#368).
const RELEASES_URL = '/releases.json';
const RELEASES_PAGE_URL =
	'https://github.com/alrayyes/pipeline-analytics/releases';

let releases = $state<Release[] | null>(null);
let failed = $state(false);

onMount(() => {
	fetch(RELEASES_URL)
		.then((res) => {
			if (!res.ok) throw new Error(`releases.json returned ${res.status}`);
			// A build that skipped the generation step falls through the SPA
			// fallback to index.html, which isn't JSON: that rejects here and
			// lands in the same "couldn't load" state as a failed request.
			return res.json();
		})
		.then((data: Release[]) => {
			releases = data;
		})
		.catch(() => {
			failed = true;
		});
});

// release-please's own body opens with a "## [version](compare-link) (date)"
// heading -- redundant here since the card header already shows the same
// version and date structurally, so it's stripped before rendering the rest.
const LEADING_VERSION_HEADING = /^##\s*\[.*?\]\(.*?\)\s*\(.*?\)\s*\n+/;

function renderBody(body: string): string {
	const withoutHeading = body.replace(LEADING_VERSION_HEADING, '');
	return DOMPurify.sanitize(marked.parse(withoutHeading, { async: false }));
}
</script>

<svelte:head>
	<title>Release history · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="text-2xl font-semibold">Release history</h1>

	{#if failed}
		<p role="alert" class="mt-6 text-destructive">
			Couldn't load releases right now. See the full list on
			<a href={RELEASES_PAGE_URL} class="underline hover:text-foreground">GitHub</a>.
		</p>
	{:else if releases === null}
		<p class="mt-6 text-muted-foreground">Loading&hellip;</p>
	{:else if releases.length === 0}
		<p class="mt-6 text-muted-foreground">No releases yet.</p>
	{:else}
		<ul class="mt-6 grid gap-3">
			{#each releases as release (release.tag)}
				<li>
					<Card size="sm">
						<CardHeader class="flex flex-row items-center justify-between">
							<CardTitle>
								<h2>
									<a
										href={release.url}
										class="hover:underline"
									>
										{release.name}
									</a>
								</h2>
							</CardTitle>
							<time
								datetime={release.date}
								class="text-sm text-muted-foreground"
							>
								<!-- A date-only string parses as UTC midnight, so format it as UTC:
								     in a timezone west of UTC the local date would be the day before. -->
								{new Date(release.date).toLocaleDateString(undefined, {
									year: 'numeric',
									month: 'long',
									day: 'numeric',
									timeZone: 'UTC',
								})}
							</time>
						</CardHeader>
						{#if release.body}
							<CardContent>
								<div
									class="prose prose-sm dark:prose-invert max-w-none prose-h3:mt-3 prose-h3:mb-1.5 prose-h3:text-[11px] prose-h3:font-semibold prose-h3:tracking-wide prose-h3:text-muted-foreground prose-h3:uppercase first:prose-h3:mt-0"
								>
									{@html renderBody(release.body)}
								</div>
							</CardContent>
						{/if}
					</Card>
				</li>
			{/each}
		</ul>
	{/if}
</main>
