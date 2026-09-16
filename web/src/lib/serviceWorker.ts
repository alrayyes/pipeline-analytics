// Registers the app-shell service worker (static/sw.js) so the app is
// installable -- guarded to production builds only, so a dev server
// doesn't fight its own hot-reloading with a service worker caching an
// old shell.
export function registerServiceWorker(): void {
	if (!import.meta.env.PROD) return;
	if (!('serviceWorker' in navigator)) return;

	navigator.serviceWorker.register('/sw.js').catch(() => {
		// Best effort -- an install prompt just won't appear if this fails.
	});
}
