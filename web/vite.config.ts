import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
			},
			// Output goes straight into the Go package that go:embeds it
			// (internal/webassets/dist), rather than web/build -- go:embed
			// can't reach outside its own package's directory tree.
			adapter: adapter({
				pages: '../internal/webassets/dist',
				assets: '../internal/webassets/dist',
				fallback: 'index.html',
			}),
		}),
	],
});
