import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
			},
			// The global stylesheet (~82 kB, ~13 kB brotli) goes into index.html
			// instead of blocking first render behind its own request. The
			// default of 0 inlines nothing; the threshold only has to clear the
			// sheet. See #527 and rules/web-performance.md.
			inlineStyleThreshold: 100_000,
			// Output goes straight into the Go package that go:embeds it
			// (internal/webassets/dist), rather than web/build -- go:embed
			// can't reach outside its own package's directory tree.
			adapter: adapter({
				pages: '../internal/webassets/dist',
				assets: '../internal/webassets/dist',
				fallback: 'index.html',
				// Writes .br and .gz beside each text file; the Go server
				// picks one by Accept-Encoding (internal/httpserver/static.go).
				precompress: true,
			}),
		}),
	],
});
