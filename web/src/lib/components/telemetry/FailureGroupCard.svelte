<script lang="ts">
import { Badge } from '#lib/components/ui/badge/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '#lib/components/ui/card/index.js';
import type { FailureGroup } from '#lib/dashboardApi.js';
import { categoryLabel } from '#lib/failureCategory.js';
import { flakyRunsHref } from '#lib/flakyRuns.js';

let { group }: { group: FailureGroup } = $props();

const count = $derived(group.pipelines.length);
</script>

<Card class="min-w-0">
	<CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2">
		<CardTitle class="contents">
			<h3 class="min-w-0 break-words">{group.step}</h3>
		</CardTitle>
		<Badge variant="secondary">{categoryLabel(group.category)}</Badge>
	</CardHeader>
	<CardContent class="grid min-w-0 gap-3 text-sm">
		<p>
			{group.occurrences}
			{group.occurrences === 1 ? 'failure' : 'failures'} in {count}
			{count === 1 ? 'pipeline' : 'pipelines'}
		</p>
		{#if group.conclusion}
			<p class="text-muted-foreground">
				Forge conclusion: <code class="font-mono text-xs">{group.conclusion}</code>
			</p>
		{/if}
		<ul class="grid gap-1">
			{#each group.pipelines as pipeline (pipeline.pipelineId)}
				<li class="min-w-0">
					<a
						href={flakyRunsHref(pipeline.pipelineId, group.step)}
						class="break-words text-foreground underline underline-offset-4"
					>
						{pipeline.pipelineName}
					</a>
				</li>
			{/each}
		</ul>
	</CardContent>
</Card>
