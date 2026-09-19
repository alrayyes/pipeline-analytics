import { plugin } from 'bun';
import { compileModule } from 'svelte/compiler';

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
	},
});
