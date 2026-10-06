<script lang="ts">
import { clearBranch, getBranch, setBranch } from '#lib/branchFilter.svelte.js';
import { ALL_BRANCHES, branchOptions } from '#lib/branchOptions.js';
import { Label } from '#lib/components/ui/label/index.js';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
} from '#lib/components/ui/select/index.js';
import {
	type BranchCount,
	fetchBranches,
	type InsightsWindow,
} from '#lib/dashboardApi.js';

let {
	window,
	onChange,
}: {
	// The window the branch list covers; omit for the server's default.
	window?: InsightsWindow;
	// Called after the choice changes, e.g. to go back to a list's first page.
	onChange?: () => void;
} = $props();

let branches = $state<BranchCount[] | undefined>();

// A failed request leaves the selector offering only "All branches": the
// views still load, so a broken branch list never takes them down with it.
$effect(() => {
	let stale = false;

	fetchBranches({ window })
		.then((body) => {
			if (!stale) branches = body.branches;
		})
		.catch(() => {
			if (!stale) branches = undefined;
		});

	return () => {
		stale = true;
	};
});

const options = $derived(branchOptions(branches, getBranch()));
const label = $derived(
	options.find((o) => o.value === (getBranch() ?? ALL_BRANCHES))?.label ??
		'All branches',
);
</script>

<div class="flex items-center gap-2">
	<Label for="branch-filter">Branch</Label>
	<Select
		type="single"
		value={getBranch() ?? ALL_BRANCHES}
		onValueChange={(value) => {
			if (!value) return;

			if (value === ALL_BRANCHES) clearBranch();
			else setBranch(value);

			onChange?.();
		}}
	>
		<SelectTrigger id="branch-filter" class="w-48">{label}</SelectTrigger>
		<SelectContent>
			{#each options as option (option.value)}
				<SelectItem value={option.value} label={option.label}>
					{option.label}
					{#if option.runCount !== undefined}
						<span class="text-muted-foreground">({option.runCount})</span>
					{/if}
				</SelectItem>
			{/each}
		</SelectContent>
	</Select>
</div>
