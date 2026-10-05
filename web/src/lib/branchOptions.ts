import type { BranchCount } from './dashboardApi.js';

// "All branches" must not collide with a real branch name; git forbids `*` in
// a ref name, so no branch can be called this.
export const ALL_BRANCHES = '*';

export interface BranchOption {
	value: string;
	label: string;
	runCount?: number;
}

// What the selector offers: all branches first, then the branches with runs in
// the window in the order the server sent (busiest first). A chosen branch the
// window has no runs for is kept, so the selector still shows what the views
// are scoped to instead of silently falling back to every branch.
export function branchOptions(
	branches: BranchCount[] | undefined,
	selected: string | null,
): BranchOption[] {
	const options: BranchOption[] = [
		{ value: ALL_BRANCHES, label: 'All branches' },
	];

	for (const { name, runCount } of branches ?? []) {
		options.push({ value: name, label: name, runCount });
	}

	if (selected && !options.some((o) => o.value === selected)) {
		options.push({ value: selected, label: selected });
	}

	return options;
}
