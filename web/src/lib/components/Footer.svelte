<script lang="ts">
import { onMount } from 'svelte';
import { page } from '$app/state';

let version = $state<string | null>(null);

onMount(() => {
	fetch('/api/version')
		.then((res) => res.json())
		.then((data: { version: string }) => {
			version = data.version;
		})
		.catch(() => {
			version = null;
		});
});
</script>

<footer>
	<p>
		pipeline-analytics
		{#if version === 'dev'}
			<span>· dev build</span>
		{:else if version}
			<a href="https://github.com/alrayyes/pipeline-analytics/releases/tag/{version}">{version}</a>
		{/if}
		{#if page.url.pathname !== '/releases'}
			· <a href="/releases">Release history</a>
		{/if}
	</p>
</footer>
