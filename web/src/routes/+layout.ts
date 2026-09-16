import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

// Embedded into the Go binary and served as static files with no Node
// server behind it -- every route renders client-side. adapter-static's
// `fallback: 'index.html'` (vite.config.ts) serves this to any unmatched
// path so the client router can take over, which is what lets a future
// dynamic route (e.g. /pipelines/[id]) work without being prerenderable at
// build time.
export const ssr = false;

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
export const load: LayoutLoad = async ({ url, fetch }) => {
	if (url.pathname === '/login') {
		return { hasRepos: false };
	}

	const res = await fetch('/api/repos');
	if (res.status === 401) {
		redirect(302, '/login');
	}

	if (!res.ok) {
		return { hasRepos: false };
	}

	const repos: unknown[] = await res.json();

	return { hasRepos: repos.length > 0 };
};
