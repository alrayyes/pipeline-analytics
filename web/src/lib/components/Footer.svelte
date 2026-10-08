<script lang="ts">
import { onMount } from 'svelte';
import { page } from '$app/state';

const repoUrl = 'https://github.com/alrayyes/pipeline-analytics';
const linkClass =
	'hover:text-foreground hover:underline focus-visible:underline';

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
	<!-- Spaced, not dot-separated: a "·" can't be kept off the end of a row
	     once the items wrap on a phone. -->
	<ul
		class="mx-auto flex max-w-4xl flex-wrap items-center justify-center gap-x-4 gap-y-1 px-4 py-4 text-xs text-muted-foreground"
	>
		<li>pipeline-analytics</li>
		{#if version === 'dev'}
			<li>dev build</li>
		{:else if version}
			<li>
				{#if page.url.pathname === '/releases'}
					<!-- The page you're on isn't a link to itself, as with the privacy link. -->
					<span>{version}</span>
				{:else}
					<a href="/releases" class={linkClass}>{version}</a>
				{/if}
			</li>
		{/if}
		{#if page.url.pathname !== '/legal'}
			<li><a href="/legal" class={linkClass}>Privacy &amp; disclaimer</a></li>
		{/if}
		<li>
			<a href="{repoUrl}/blob/main/LICENSE" class={linkClass}>AGPL-3.0</a>
		</li>
		<li>
			<a href={repoUrl} class="inline-flex items-center gap-1 {linkClass}">
				<!-- Simple Icons' GitHub mark (CC0), vendored; the label names the link. -->
				<svg
					viewBox="0 0 24 24"
					class="size-3.5 fill-current"
					aria-hidden="true"
				>
					<path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
				</svg>
				GitHub
			</a>
		</li>
	</ul>
</footer>
