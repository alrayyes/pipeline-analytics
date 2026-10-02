<script lang="ts">
import {
	ToggleGroup,
	ToggleGroupItem,
} from '$lib/components/ui/toggle-group/index.js';
import type { InsightsWindow } from '$lib/dashboardApi.js';

let {
	value,
	onChange,
}: { value: InsightsWindow; onChange: (next: InsightsWindow) => void } =
	$props();

const OPTIONS: { value: InsightsWindow; short: string; label: string }[] = [
	{ value: '24h', short: '24h', label: '24 hours' },
	{ value: '7d', short: '7d', label: '7 days' },
	{ value: '30d', short: '30d', label: '30 days' },
];
</script>

<!-- radiogroup parent for the radio items: see ForgeFilter.svelte. -->
<div role="radiogroup" aria-label="Time window">
	<ToggleGroup
		type="single"
		variant="outline"
		{value}
		onValueChange={(next) => {
			if (next) onChange(next as InsightsWindow);
		}}
	>
		{#each OPTIONS as option (option.value)}
			<ToggleGroupItem value={option.value} aria-label={option.label}>
				{option.short}
			</ToggleGroupItem>
		{/each}
	</ToggleGroup>
</div>
