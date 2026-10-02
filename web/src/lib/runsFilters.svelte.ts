import type { RunStatusFilter } from './dashboardApi.js';

// Mirrors the Pipelines and Repos pages' page size (#165/#166).
export const RUNS_PAGE_SIZE = 20;

// Page-local view state, not an account setting: which status tab the runs
// list is on and how far into it the reader has paged. Module-level runes so
// the choice survives a navigation away and back within a session, but
// nothing is sent to the server or cached (unlike telemetryWindow).
let status = $state<RunStatusFilter>('all');
let offset = $state(0);

export function getStatus(): RunStatusFilter {
	return status;
}

export function getOffset(): number {
	return offset;
}

// A new status is a new list, so it starts at the top; re-selecting the
// status already shown leaves the reader where they are.
export function setStatus(next: RunStatusFilter): void {
	if (next === status) return;

	status = next;
	offset = 0;
}

export function nextPage(): void {
	offset += RUNS_PAGE_SIZE;
}

export function previousPage(): void {
	offset = Math.max(0, offset - RUNS_PAGE_SIZE);
}

export function resetRunsFilters(): void {
	status = 'all';
	offset = 0;
}
