<script lang="ts">
import DOMPurify from 'dompurify';
import { marked } from 'marked';
import { onMount } from 'svelte';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';

interface Release {
	tag_name: string;
	name: string | null;
	html_url: string;
	published_at: string;
	body: string | null;
}

const GITHUB_RELEASES_URL =
	'https://api.github.com/repos/alrayyes/pipeline-analytics/releases';
const RELEASES_PAGE_URL =
	'https://github.com/alrayyes/pipeline-analytics/releases';

let releases = $state<Release[] | null>(null);
let failed = $state(false);

onMount(() => {
	fetch(GITHUB_RELEASES_URL, {
		headers: { Accept: 'application/vnd.github+json' },
	})
		.then((res) => {
			if (!res.ok) throw new Error(`GitHub API returned ${res.status}`);
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
			{#each releases as release (release.tag_name)}
				<li>
					<Card size="sm">
						<CardHeader class="flex flex-row items-center justify-between">
							<CardTitle>
								<h2>
									<a
										href={release.html_url}
										class="hover:underline"
									>
										{release.name || release.tag_name}
									</a>
								</h2>
							</CardTitle>
							<time
								datetime={release.published_at}
								class="text-sm text-muted-foreground"
							>
								{new Date(release.published_at).toLocaleDateString(undefined, {
									year: 'numeric',
									month: 'long',
									day: 'numeric',
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
