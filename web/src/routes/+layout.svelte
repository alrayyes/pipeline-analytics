<script lang="ts">
import '../app.css';
import { page } from '$app/state';
import favicon from '$lib/assets/favicon.svg';
import Footer from '$lib/components/Footer.svelte';
import Nav from '$lib/components/Nav.svelte';

let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if page.url.pathname !== '/login'}
	<Nav />
{/if}

{@render children()}

<Footer />
