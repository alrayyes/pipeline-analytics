<script lang="ts">
import { onMount } from 'svelte';

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
</script>

<svelte:head>
	<title>Release history</title>
</svelte:head>

<main>
	<h1>Release history</h1>

	{#if failed}
		<p>
			Couldn't load releases right now. See the full list on
			<a href={RELEASES_PAGE_URL}>GitHub</a>.
		</p>
	{:else if releases === null}
		<p>Loading&hellip;</p>
	{:else if releases.length === 0}
		<p>No releases yet.</p>
	{:else}
		<ul>
			{#each releases as release (release.tag_name)}
				<li>
					<h2><a href={release.html_url}>{release.name || release.tag_name}</a></h2>
					<time datetime={release.published_at}>
						{new Date(release.published_at).toLocaleDateString()}
					</time>
					{#if release.body}
						<p>{release.body}</p>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</main>
