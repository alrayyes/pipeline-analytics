<script lang="ts">
import { buttonVariants } from '#lib/components/ui/button/index.js';
import { errorView } from '#lib/errors.js';
import { cn } from '#lib/utils.js';
import { page } from '$app/state';

// Only the status picks the wording. page.error.message is deliberately not
// rendered: handleError already made it generic, and this view does not
// depend on that staying true (#576).
const view = $derived(errorView(page.status));
</script>

<svelte:head>
	<title>{view.title} · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto flex max-w-xl flex-col items-start gap-4 px-4 py-16">
	<span
		class="rounded-md border border-foreground px-2 py-0.5 text-xs font-medium"
	>
		{view.chip}
	</span>
	<h1 class="text-3xl font-semibold">{view.title}</h1>
	<p class="text-muted-foreground">{view.explanation}</p>

	<div class="mt-2 grid w-full gap-2 sm:grid-cols-2">
		<a href="/" class={cn(buttonVariants(), 'w-full')}>Back to the overview</a>
		<a href="/pipelines" class={cn(buttonVariants({ variant: 'outline' }), 'w-full')}>
			Browse pipelines
		</a>
	</div>
</main>
