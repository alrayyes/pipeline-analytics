<script lang="ts">
import {
	ToggleGroup,
	ToggleGroupItem,
} from '$lib/components/ui/toggle-group/index.js';
import {
	type ForgeFilter,
	getForgeFilter,
	setForgeFilter,
} from '$lib/forgeFilter.svelte.js';

const OPTIONS: { value: ForgeFilter; label: string }[] = [
	{ value: 'all', label: 'All' },
	{ value: 'github', label: 'GitHub' },
	{ value: 'forgejo', label: 'Forgejo' },
];
</script>

<!--
bits-ui's ToggleGroup root always renders role="group" (its own computed
props win the merge over anything passed in), but type="single" makes each
item role="radio" -- which needs a radiogroup ancestor per ARIA, or axe's
aria-required-parent flags it. Supplying that from an outer element here,
since overriding the primitive's own role isn't possible through its props.
-->
<div role="radiogroup" aria-label="Filter by forge">
	<ToggleGroup
		type="single"
		variant="outline"
		value={getForgeFilter()}
		onValueChange={(value) => {
			if (value) setForgeFilter(value as ForgeFilter);
		}}
	>
		{#each OPTIONS as option (option.value)}
			<ToggleGroupItem value={option.value} aria-label={option.label}>
				{option.label}
			</ToggleGroupItem>
		{/each}
	</ToggleGroup>
</div>
