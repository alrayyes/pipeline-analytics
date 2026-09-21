import { redirect } from '@sveltejs/kit';
import { fetchSettings } from '$lib/settingsSync.js';
import type { LayoutLoad } from './$types';

// Embedded into the Go binary and served as static files with no Node
// server behind it -- every route renders client-side. adapter-static's
// `fallback: 'index.html'` (vite.config.ts) serves this to any unmatched
// path so the client router can take over, which is what lets a future
// dynamic route (e.g. /pipelines/[id]) work without being prerenderable at
// build time.
export const ssr = false;

// #251: this load() gates the ENTIRE client-side render -- ssr is off, so
// nothing (not even the nav) mounts until it resolves. A `fetch()` with no
// timeout is an unbounded wait on the caller's side of that gate: under
// heavy concurrent load (confirmed via a captured trace: repeated parallel
// e2e workers, each running its own full server, occasionally left a
// single GET taking far longer than normal, with no server error and no
// dropped connection -- just no response for tens of seconds) the whole
// page stayed blank forever, because there was nothing to time out on.
// Bounding both calls and falling back to a safe default on abort -- same
// as fetchSettings already does for any other failure -- means a stalled
// backend degrades this navigation instead of freezing it.
const REQUEST_TIMEOUT_MS = 8_000;

// The server gates every route except login/registration with a session
// check that returns a JSON 401, not a server-side redirect (design.md's
// SPA-fallback decision) -- this load function is the client-side redirect
// that 401 is meant to trigger. It reads `url`, so SvelteKit reruns it on
// every navigation, catching a session that expired mid-visit too.
//
// The same response also answers whether any repo is registered at all
// (issue #71) -- the nav and the Pipelines empty state both need that to
// pick the right copy/CTA, and this call already fires on every navigation
// for the 401 check above, so reading its body costs nothing extra.
//
// The settings fetch (persist-account-settings/design.md's "folds into the
// existing auth-gate request") runs in parallel with the repos call via
// Promise.all, not after it -- a second sequential round trip here would
// delay every navigation just to seed stores that already have a
// localStorage-cached value to paint from in the meantime. A failed
// settings fetch resolves to null (fetchSettings' own doc comment), not an
// error -- the root layout's init calls fall back to that cache.
export const load: LayoutLoad = async ({ url, fetch }) => {
	if (url.pathname === '/login') {
		return { hasRepos: false, settings: null };
	}

	const signal = AbortSignal.timeout(REQUEST_TIMEOUT_MS);

	// limit=1: this call only needs to know whether any repo exists at all,
	// not fetch the page a user is looking at. A timed-out/aborted repos
	// check can't tell a real 401 from a stalled backend, so it falls back
	// to `hasRepos: false` rather than guessing -- the same trade-off
	// fetchSettings already makes for its own failures, and the next real
	// navigation re-checks properly either way.
	const [reposRes, settings] = await Promise.all([
		fetch('/api/repos?limit=1', { signal }).catch(() => null),
		fetchSettings(fetch, signal),
	]);

	if (reposRes?.status === 401) {
		redirect(302, '/login');
	}

	if (!reposRes?.ok) {
		return { hasRepos: false, settings };
	}

	const body: { repos: unknown[] } = await reposRes.json();

	return { hasRepos: body.repos.length > 0, settings };
};
