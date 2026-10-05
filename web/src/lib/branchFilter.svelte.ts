// The branch the telemetry views are scoped to, or null for every branch.
// View state shared by the overview, root-cause, flaky and runs views, so a
// branch chosen on one is the branch the next opens with. Not an account
// setting and not cached: it lasts the visit, and a reload goes back to
// every branch.
let branch = $state<string | null>(null);

export function getBranch(): string | null {
	return branch;
}

export function setBranch(next: string): void {
	branch = next === '' ? null : next;
}

export function clearBranch(): void {
	branch = null;
}
