// Re-running a run on its forge (#481). The server does the work and answers
// 202 when the forge accepts the request, not when the run has finished, so
// the UI says "asked", never "done". Failures keep the server's `code`, and
// the wording is decided here so each outcome reads differently.

import type { RunSummary } from './dashboardApi.js';

export type RunActionKind = 'rerun';

export type RunActionResult =
	| { ok: true }
	| { ok: false; status: number; code: string };

export async function rerunRun(
	runId: string,
	fetchFn: typeof fetch = fetch,
): Promise<RunActionResult> {
	try {
		const res = await fetchFn(`/api/runs/${encodeURIComponent(runId)}/rerun`, {
			method: 'POST',
		});
		if (res.ok) return { ok: true };

		const body = (await res.json().catch(() => ({}))) as { code?: string };

		return { ok: false, status: res.status, code: body.code ?? 'unknown' };
	} catch {
		return { ok: false, status: 0, code: 'network' };
	}
}

const MESSAGES: Record<string, string> = {
	forbidden:
		'The saved token needs Actions write permission to re-run a run on GitHub.',
	not_actionable: 'This run can only be re-run once it has finished.',
	unsupported: 'This forge doesn’t support re-running from here.',
	unreachable: 'The forge didn’t answer. Try again in a moment.',
	not_found: 'Run not found.',
	network: 'Could not reach the server.',
};

export function actionMessage(
	_kind: RunActionKind,
	result: Extract<RunActionResult, { ok: false }>,
): string {
	return MESSAGES[result.code] ?? 'Could not re-run this run.';
}

// Which runs may be re-run is the server's call (`actions` on the run); the
// UI only renders it.
export function offersRerun(run: Pick<RunSummary, 'actions'>): boolean {
	return run.actions.includes('rerun');
}
