<script lang="ts">
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '#lib/components/ui/card/index.js';
import type { FlakyStep } from '#lib/dashboardApi.js';
import { flakyRunsHref } from '#lib/flakyRuns.js';
import FlakeMatrix from './FlakeMatrix.svelte';
import StatusBadge from './StatusBadge.svelte';

let { step }: { step: FlakyStep } = $props();
</script>

<Card class="min-w-0">
	<CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2">
		<CardTitle class="contents">
			<h3 class="min-w-0 break-words">{step.name}</h3>
		</CardTitle>
		<StatusBadge tone="flaky" label="Flaky" />
	</CardHeader>
	<CardContent class="grid min-w-0 gap-3 text-sm">
		<p class="text-muted-foreground">{step.pipelineName}</p>
		<FlakeMatrix {step} />
		<a
			href={flakyRunsHref(step.pipelineId, step.name)}
			class="w-fit text-foreground underline underline-offset-4"
		>
			View failed runs
		</a>
	</CardContent>
</Card>
