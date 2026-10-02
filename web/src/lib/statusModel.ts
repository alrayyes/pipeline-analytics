// Maps the API's normalized outcome to the tones the telemetry views draw
// and the text beside each. What a forge's status or conclusion means is
// decided by the server (metrics.OutcomeOf) and arrives as `outcome`; this
// file only decides how an outcome looks. Colour is never the only signal:
// every tone has a label a component renders next to its icon.

import type { Outcome } from './dashboardApi.js';

export type Tone = 'pass' | 'fail' | 'running' | 'flaky' | 'neutral';

const TONES: Record<Outcome, Tone> = {
	passed: 'pass',
	failed: 'fail',
	running: 'running',
	// Work still to come reads like running: it is about to be.
	queued: 'running',
	cancelled: 'neutral',
	skipped: 'neutral',
	unknown: 'neutral',
};

const LABELS: Record<Exclude<Outcome, 'unknown'>, string> = {
	passed: 'Passed',
	failed: 'Failed',
	running: 'Running',
	queued: 'Queued',
	cancelled: 'Cancelled',
	skipped: 'Skipped',
};

export function outcomeTone(outcome: Outcome): Tone {
	return TONES[outcome];
}

// An outcome the server couldn't classify shows the forge's own text (the
// conclusion, else the status) rather than a blank or a guess.
export function outcomeLabel(outcome: Outcome, fallback?: string): string {
	return outcome === 'unknown' ? (fallback ?? 'Unknown') : LABELS[outcome];
}

// Work still to come: a list showing it should keep refreshing.
export function isPending(outcome: Outcome): boolean {
	return outcome === 'running' || outcome === 'queued';
}
