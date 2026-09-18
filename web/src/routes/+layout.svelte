<script lang="ts">
import { onMount } from 'svelte';
import '../app.css';
import { page } from '$app/state';
import favicon from '$lib/assets/favicon.svg';
import Footer from '$lib/components/Footer.svelte';
import Nav from '$lib/components/Nav.svelte';
import { initForgeFilter } from '$lib/forgeFilter.svelte.js';
import { initPipelinesFilters } from '$lib/pipelinesFilters.svelte.js';
import { registerServiceWorker } from '$lib/serviceWorker.js';
import { initTheme } from '$lib/theme.svelte.js';

let { data, children } = $props();

// Re-runs on every navigation, not just the first mount -- +layout.ts's
// load() (and the settings fetch it now folds in) already reruns per
// navigation for the session check, and persist-account-settings/
// design.md's own stated non-goal is exactly this: "a page reload or
// navigation is sufficient to pick up a change made on another device",
// not real-time sync.
//
// Svelte also reschedules this effect on its own -- e.g. right after a
// settings PATCH fires from a store's setter -- with the very same `data`
// object SvelteKit handed it for the current navigation. Reapplying
// unconditionally there would stomp the local change that PATCH is still
// in flight for, resetting the UI back to the stale server value before
// the response even lands. Guarding on `data`'s own identity -- reference-
// equal only across those spurious re-fires, never across a real
// navigation -- reconciles once per navigation and no-ops on the rest.
let reconciledData: typeof data | undefined;

$effect(() => {
	if (data === reconciledData) return;
	reconciledData = data;

	const cleanup = initTheme(data.settings?.theme);
	initForgeFilter(data.settings?.forgeFilter);
	initPipelinesFilters(data.settings ?? undefined);

	return cleanup;
});

onMount(registerServiceWorker);
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if page.url.pathname !== '/login'}
	<Nav hasRepos={data.hasRepos} />
{/if}

{@render children()}

<Footer />
