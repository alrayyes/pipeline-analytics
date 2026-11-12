// What a visitor sees when something fails (#576, rules/error-pages.md):
// plain words, no status code as the headline, and nothing about what broke.
// The detail goes to the log, not the page.

import type { CaughtError } from '@sveltejs/kit/hooks';

export interface ErrorView {
	chip: string;
	title: string;
	explanation: string;
}

export function errorView(status: number): ErrorView {
	if (status === 404) {
		return {
			chip: 'Not found',
			title: 'Page not found',
			explanation: "That page doesn't exist, or it has moved.",
		};
	}

	return {
		chip: 'Server error',
		title: 'Something went wrong',
		explanation:
			"We couldn't load this page. Try again, or head back to the overview.",
	};
}

// Shared by hooks.server.ts and hooks.client.ts. The returned message is the
// only thing `page.error` carries to the view, so it is generic whatever was
// thrown. A missing page is the visitor's typo, not a fault, so it isn't
// logged, whether the router found no route (framework) or a load function
// threw error(404) (app).
export function shapeError(input: CaughtError): { message: string } {
	if (input.kind !== 'unknown' && input.error.status === 404) {
		return { message: 'Page not found' };
	}

	console.error(input.error);

	return { message: 'Something went wrong' };
}
