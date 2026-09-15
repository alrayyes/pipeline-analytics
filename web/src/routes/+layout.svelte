<script lang="ts">
import favicon from '$lib/assets/favicon.svg';
import Footer from '$lib/components/Footer.svelte';

let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<Footer />
