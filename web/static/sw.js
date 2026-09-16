// Precaches only the app shell (the index document adapter-static's SPA
// fallback always serves), not any API response -- this isn't an
// offline-first app (every page here calls a live API), so caching /api/*
// would mean serving stale run/pipeline data with no way to know it's
// stale. This exists solely to satisfy installability: a PWA install
// prompt requires an active service worker with a fetch handler.
const CACHE = 'pipeline-analytics-shell-v1';
const SHELL = ['/'];

self.addEventListener('install', (event) => {
	event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(SHELL)));
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(
					keys.filter((key) => key !== CACHE).map((key) => caches.delete(key)),
				),
			)
			.then(() => self.clients.claim()),
	);
});

// Network-first for navigations, falling back to the cached shell only if
// the network is actually unreachable -- everything else (including every
// /api/* call) is untouched, straight through to the network.
self.addEventListener('fetch', (event) => {
	if (event.request.mode !== 'navigate') return;

	event.respondWith(fetch(event.request).catch(() => caches.match('/')));
});
