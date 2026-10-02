<script lang="ts">
import type { RunStep } from '$lib/dashboardApi.js';
import { stageSegments, stageSummary } from '$lib/stageProgress.js';
import type { Tone } from '$lib/statusModel.js';

let { steps }: { steps: RunStep[] } = $props();

const segments = $derived(stageSegments(steps));

const BAR_CLASSES: Record<Tone, string> = {
	pass: 'bg-green-600 dark:bg-green-500',
	fail: 'bg-destructive',
	running: 'bg-sky-600 dark:bg-sky-500 motion-safe:animate-pulse',
	flaky: 'bg-amber-600 dark:bg-amber-500',
	neutral: 'bg-muted',
};
</script>

<div class="grid min-w-0 gap-2">
	<p class="text-sm text-muted-foreground">{stageSummary(steps)}</p>
	<ol class="flex min-w-0 gap-1">
		{#each segments as segment, i (i)}
			<li class="flex min-w-0 flex-1 flex-col gap-1">
				<div class="h-2 rounded-full {BAR_CLASSES[segment.tone]}"></div>
				<span class="truncate text-xs text-muted-foreground" aria-hidden="true">
					{segment.name}
				</span>
				<span class="sr-only">{segment.name}: {segment.label}</span>
			</li>
		{/each}
	</ol>
</div>
