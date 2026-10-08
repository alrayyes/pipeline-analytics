// Fails the lighthouse job when a page was audited as something else. A
// signed-out browser is sent to /login, which scores well and would hide the
// fact that the page behind it was never looked at (#520). Run by
// scripts/lighthouse.sh after `lhci autorun`.

import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

interface Report {
	requestedUrl: string;
	finalDisplayedUrl: string;
}

const trimmed = (url: string): string => url.replace(/\/$/, '');

export function findRedirected(reports: Report[]): string[] {
	return reports
		.filter((r) => trimmed(r.requestedUrl) !== trimmed(r.finalDisplayedUrl))
		.map((r) => `${r.requestedUrl} ended on ${r.finalDisplayedUrl}`);
}

if (import.meta.main) {
	const dir = process.argv[2] ?? '.lighthouseci';
	const reports = readdirSync(dir)
		.filter((name) => /^lhr-.*\.json$/.test(name))
		.map((name) => JSON.parse(readFileSync(join(dir, name), 'utf8')) as Report);

	if (reports.length === 0) {
		console.error(`check-lighthouse-reports: no reports in ${dir}`);
		process.exit(1);
	}

	const redirected = findRedirected(reports);
	for (const line of redirected) {
		console.error(`check-lighthouse-reports: ${line}`);
	}
	if (redirected.length > 0) {
		process.exit(1);
	}
	console.log(`check-lighthouse-reports: ${reports.length} pages as requested`);
}
