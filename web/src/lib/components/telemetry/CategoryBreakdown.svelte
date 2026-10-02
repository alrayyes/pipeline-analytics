<script lang="ts">
import type { CategoryCount } from '$lib/dashboardApi.js';
import { categoryLabel } from '$lib/failureCategory.js';
import { formatRate } from '$lib/format.js';

let { categories }: { categories: CategoryCount[] } = $props();

// Presentation only: the bar is decorative, the rows below carry the figures.
// Tones differ in lightness and the rows repeat each swatch, so no meaning
// rides on colour alone.
const TONES = [
	'bg-foreground',
	'bg-foreground/70',
	'bg-foreground/45',
	'bg-foreground/30',
	'bg-foreground/15',
];

function tone(index: number): string {
	return TONES[index % TONES.length];
}
</script>

{#if categories.length > 0}
	<section aria-labelledby="failure-categories-heading" class="mt-6">
		<h2 id="failure-categories-heading" class="text-lg font-semibold">
			Failure categories
		</h2>
		<div
			aria-hidden="true"
			class="mt-3 flex h-3 w-full overflow-hidden rounded-full border"
		>
			{#each categories as item, i (item.category)}
				<div
					class="{tone(i)} h-full border-r border-background last:border-r-0"
					style:width="{item.share * 100}%"
				></div>
			{/each}
		</div>
		<ul class="mt-3 grid gap-2">
			{#each categories as item, i (item.category)}
				<li class="flex min-w-0 items-center gap-3 text-sm">
					<span
						aria-hidden="true"
						class="{tone(i)} size-3 shrink-0 rounded-sm border"
					></span>
					<span class="min-w-0 flex-1 truncate">{categoryLabel(item.category)}</span>
					<span class="shrink-0 font-medium">{formatRate(item.share)}</span>
					<span class="shrink-0 text-muted-foreground">
						{item.occurrences}
						{item.occurrences === 1 ? 'failure' : 'failures'}
					</span>
				</li>
			{/each}
		</ul>
	</section>
{/if}
