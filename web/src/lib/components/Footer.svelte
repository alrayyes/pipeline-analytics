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

<footer class="border-t">
	<p class="mx-auto max-w-4xl px-4 py-4 text-center text-xs text-muted-foreground">
		pipeline-analytics
		{#if version === 'dev'}
			<span>· dev build</span>
		{:else if version}
			·
			<a
				href="https://github.com/alrayyes/pipeline-analytics/releases/tag/{version}"
				class="hover:text-foreground hover:underline"
			>
				{version}
			</a>
		{/if}
		{#if page.url.pathname !== '/releases'}
			·
			<a href="/releases" class="hover:text-foreground hover:underline">Release history</a>
		{/if}
		{#if page.url.pathname !== '/legal'}
			·
			<a href="/legal" class="hover:text-foreground hover:underline">Privacy &amp; disclaimer</a>
		{/if}
	</p>
</footer>
