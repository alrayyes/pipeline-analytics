<script lang="ts">
import type { FlakyStep } from '$lib/dashboardApi.js';
import { flakeCells, flakeSummary } from '$lib/flakeMatrix.js';
import type { Tone } from '$lib/statusModel.js';

let { step }: { step: FlakyStep } = $props();

const cells = $derived(flakeCells(step.recentOutcomes));
const summary = $derived(flakeSummary(step));

const CELL_CLASSES: Record<Tone, string> = {
	pass: 'bg-green-600 dark:bg-green-500',
	fail: 'bg-destructive text-white',
	running: 'bg-sky-600 dark:bg-sky-500',
	flaky: 'bg-amber-600 dark:bg-amber-500',
	neutral: 'border border-border bg-muted',
};
</script>

<div class="grid min-w-0 gap-2">
	<div role="img" aria-label={summary} class="flex min-w-0 flex-wrap gap-1">
		{#each cells as cell, i (i)}
			<span
				aria-hidden="true"
				class="flex size-4 items-center justify-center rounded-sm text-[10px] leading-none font-bold {CELL_CLASSES[cell.tone]}"
			>
				{#if cell.tone === 'fail'}×{/if}
			</span>
		{/each}
	</div>
	<p class="text-sm text-muted-foreground">{summary}</p>
</div>
