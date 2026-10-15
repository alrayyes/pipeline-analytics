import { GlobalRegistrator } from '@happy-dom/global-registrator';

// A DOM for the component tests (*.dom.test.ts). Importing this registers one
// before anything else in the test file loads, since Svelte reads `document`
// when a component mounts. A test file opts in by importing it first, so the
// plain unit tests keep running without a `window`, as they do in production
// for the modules that check for one.
GlobalRegistrator.register({ url: 'http://localhost/' });

export function removeDom(): void {
	void GlobalRegistrator.unregister();
}
