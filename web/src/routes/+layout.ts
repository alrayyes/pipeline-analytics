// Embedded into the Go binary and served as static files with no Node
// server behind it -- every route renders client-side. adapter-static's
// `fallback: 'index.html'` (vite.config.ts) serves this to any unmatched
// path so the client router can take over, which is what lets a future
// dynamic route (e.g. /pipelines/[id]) work without being prerenderable at
// build time.
export const ssr = false;
