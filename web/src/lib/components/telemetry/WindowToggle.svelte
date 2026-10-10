<script lang="ts">
import {
	ToggleGroup,
	ToggleGroupItem,
} from '#lib/components/ui/toggle-group/index.js';
import type { InsightsWindow } from '#lib/dashboardApi.js';

let {
	value,
	onChange,
}: { value: InsightsWindow | null; onChange: (next: InsightsWindow) => void } =
	$props();

const OPTIONS: { value: InsightsWindow; label: string }[] = [
	{ value: '24h', label: '24 hours' },
	{ value: '7d', label: '7 days' },
	{ value: '30d', label: '30 days' },
];
</script>

<!-- radiogroup parent for the radio items: see ForgeFilter.svelte. -->
<div role="radiogroup" aria-label="Time window">
	<ToggleGroup
		type="single"
		variant="outline"
		value={value ?? ''}
		onValueChange={(next) => {
			if (next) onChange(next as InsightsWindow);
		}}
	>
		{#each OPTIONS as option (option.value)}
			<!-- Visible text is the accessible name: an aria-label that left out
			the visible "24h" fails label-content-name-mismatch. -->
			<ToggleGroupItem value={option.value}>
				{option.label}
			</ToggleGroupItem>
		{/each}
	</ToggleGroup>
</div>
