// One mapping from a forge's run/step status and conclusion to the four
// tones the telemetry views draw (pass, fail, running, flaky) plus a neutral
// one, and to the text beside each. Colour is never the only signal: every
// tone has a label a component renders next to its icon.

export type Tone = 'pass' | 'fail' | 'running' | 'flaky' | 'neutral';

const FAILED = new Set(['failure', 'timed_out']);
const RUNNING = new Set(['in_progress', 'queued', 'waiting', 'pending']);

export function statusTone(status: string, conclusion?: string): Tone {
	// A conclusion wins: a stale in_progress status next to a concluded
	// run means the status hasn't caught up, not that it's still going.
	if (conclusion === 'success') return 'pass';
	if (conclusion && FAILED.has(conclusion)) return 'fail';
	if (conclusion) return 'neutral';

	return RUNNING.has(status) ? 'running' : 'neutral';
}

const CONCLUSION_LABELS: Record<string, string> = {
	success: 'Passed',
	failure: 'Failed',
	timed_out: 'Timed out',
	cancelled: 'Cancelled',
	skipped: 'Skipped',
};

const STATUS_LABELS: Record<string, string> = {
	in_progress: 'Running',
	queued: 'Queued',
	waiting: 'Queued',
	pending: 'Queued',
};

export function statusLabel(status: string, conclusion?: string): string {
	if (conclusion && CONCLUSION_LABELS[conclusion]) {
		return CONCLUSION_LABELS[conclusion];
	}

	return STATUS_LABELS[status] ?? status;
}
