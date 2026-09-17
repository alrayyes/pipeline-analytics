<script lang="ts">
import { onMount } from 'svelte';
import '../app.css';
import { page } from '$app/state';
import favicon from '$lib/assets/favicon.svg';
import Footer from '$lib/components/Footer.svelte';
import Nav from '$lib/components/Nav.svelte';
import { initForgeFilter } from '$lib/forgeFilter.svelte.js';
import { registerServiceWorker } from '$lib/serviceWorker.js';
import { initTheme } from '$lib/theme.svelte.js';

let { data, children } = $props();

onMount(initTheme);
onMount(initForgeFilter);
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
