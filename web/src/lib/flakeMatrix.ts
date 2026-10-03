// What the flaky view draws for one step: a cell per recent result and the
// sentence that says the same thing without the grid. Which results count as
// a failure is the server's call (the outcome); this only counts and labels
// what it was sent.

import type { FlakyStep, Outcome } from './dashboardApi.js';
import { formatRate } from './format.js';
import { outcomeLabel, outcomeTone, type Tone } from './statusModel.js';

export interface FlakeCell {
	tone: Tone;
	label: string;
}

export function flakeCells(outcomes: Outcome[]): FlakeCell[] {
	return outcomes.map((outcome) => ({
		tone: outcomeTone(outcome),
		label: outcomeLabel(outcome),
	}));
}

function runs(count: number): string {
	return `${count} ${count === 1 ? 'run' : 'runs'}`;
}

// The matrix's text equivalent (the dashboard-ui spec's "Matrix has a text
// equivalent"): how many of the shown runs failed, then the server's rate
// over every run in the window, so a reader who never sees the grid gets
// the flakes and the total.
export function flakeSummary(step: FlakyStep): string {
	const failed = step.recentOutcomes.filter(
		(outcome) => outcome === 'failed',
	).length;

	return `${failed} of the last ${runs(step.recentOutcomes.length)} failed. ${formatRate(step.flakeRate)} failure rate over ${runs(step.runCount)}.`;
}
