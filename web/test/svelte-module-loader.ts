import { plugin } from 'bun';
import { compile, compileModule } from 'svelte/compiler';

const transpiler = new Bun.Transpiler({ loader: 'ts' });

plugin({
	name: 'svelte-module-compiler',
	setup(build) {
		build.onLoad({ filter: /\.svelte\.(js|ts)$/ }, async (args) => {
			const source = await Bun.file(args.path).text();
			const stripped = args.path.endsWith('.ts')
				? transpiler.transformSync(source)
				: source;
			const { js } = compileModule(stripped, {
				generate: 'client',
				filename: args.path,
			});

			return { contents: js.code, loader: 'js' };
		});

		// A component, for the tests that mount one in a DOM (test/dom.ts). The
		// filter ends at `.svelte`, so a `.svelte.ts` module still goes through
		// compileModule above. Styles are injected so no CSS file is needed.
		build.onLoad({ filter: /\.svelte$/ }, async (args) => {
			const source = await Bun.file(args.path).text();
			const { js } = compile(source, {
				generate: 'client',
				filename: args.path,
				css: 'injected',
			});

			return { contents: js.code, loader: 'js' };
		});
	},
});
