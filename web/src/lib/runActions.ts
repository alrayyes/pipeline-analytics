// Re-running or cancelling a run on its forge. The server does the work and answers
// 202 when the forge accepts the request, not when the run has finished, so
// the UI says "asked", never "done". Failures keep the server's `code`, and
// the wording is decided here so each outcome reads differently.

import type { RunSummary } from './dashboardApi.js';

export type RunActionKind = 'rerun' | 'cancel';

export type RunActionResult =
	| { ok: true }
	| { ok: false; status: number; code: string };

async function postAction(
	kind: RunActionKind,
	runId: string,
	fetchFn: typeof fetch,
): Promise<RunActionResult> {
	try {
		const res = await fetchFn(
			`/api/runs/${encodeURIComponent(runId)}/${kind}`,
			{
				method: 'POST',
			},
		);
		if (res.ok) return { ok: true };

		const body = (await res.json().catch(() => ({}))) as { code?: string };

		return { ok: false, status: res.status, code: body.code ?? 'unknown' };
	} catch {
		return { ok: false, status: 0, code: 'network' };
	}
}

export function rerunRun(
	runId: string,
	fetchFn: typeof fetch = fetch,
): Promise<RunActionResult> {
	return postAction('rerun', runId, fetchFn);
}

export function cancelRun(
	runId: string,
	fetchFn: typeof fetch = fetch,
): Promise<RunActionResult> {
	return postAction('cancel', runId, fetchFn);
}

const VERBS: Record<RunActionKind, string> = {
	rerun: 're-run',
	cancel: 'cancel',
};

const NOT_ACTIONABLE: Record<RunActionKind, string> = {
	rerun: 'This run can only be re-run once it has finished.',
	cancel: 'This run has already finished, so there is nothing to cancel.',
};

export function actionMessage(
	kind: RunActionKind,
	result: Extract<RunActionResult, { ok: false }>,
): string {
	switch (result.code) {
		case 'forbidden':
			return `The saved token needs Actions write permission to ${VERBS[kind]} a run on GitHub. Replace it in Settings with one that has it.`;
		case 'not_actionable':
			return NOT_ACTIONABLE[kind];
		case 'unsupported':
			return `This forge doesn’t support ${kind === 'rerun' ? 're-running' : 'cancelling'} from here.`;
		case 'unreachable':
			return 'The forge didn’t answer. Try again in a moment.';
		case 'not_found':
			return 'Run not found.';
		case 'network':
			return 'Could not reach the server.';
		default:
			return `Could not ${VERBS[kind]} this run.`;
	}
}

// Which runs may be re-run is the server's call (`actions` on the run); the
// UI only renders it.
export function offersRerun(run: Pick<RunSummary, 'actions'>): boolean {
	return run.actions.includes('rerun');
}

export function offersCancel(run: Pick<RunSummary, 'actions'>): boolean {
	return run.actions.includes('cancel');
}
