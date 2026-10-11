import { afterEach, describe, expect, spyOn, test } from 'bun:test';
import { errorView, shapeError } from './errors.js';

describe('errorView', () => {
	test('says a missing page is missing, in plain words', () => {
		const view = errorView(404);

		expect(view.title).toBe('Page not found');
		expect(view.chip).toBe('Not found');
		expect(view.explanation).toBe("That page doesn't exist, or it has moved.");
	});

	test('gives every other status the generic message', () => {
		for (const status of [400, 500, 502, 503]) {
			const view = errorView(status);

			expect(view.title).toBe('Something went wrong');
			expect(view.chip).toBe('Server error');
			expect(view.explanation).toBe(
				"We couldn't load this page. Try again, or head back to the overview.",
			);
		}
	});

	test('never puts a status code in the headline or chip', () => {
		for (const status of [404, 500]) {
			const view = errorView(status);

			expect(view.title).not.toMatch(/\d/);
			expect(view.chip).not.toMatch(/\d/);
		}
	});
});

describe('shapeError', () => {
	const logged = spyOn(console, 'error').mockImplementation(() => {});

	afterEach(() => logged.mockClear());

	test('returns the generic message, whatever was thrown', () => {
		const shaped = shapeError({
			error: new Error('ENOENT /srv/app/secret.db at drivers/s3.go:318'),
			status: 500,
			message: 'ENOENT /srv/app/secret.db',
		});

		expect(shaped).toEqual({ message: 'Something went wrong' });
	});

	test('logs the detail for the operator', () => {
		const error = new Error('boom');

		shapeError({ error, status: 500, message: 'boom' });

		expect(logged).toHaveBeenCalledTimes(1);
		expect(logged.mock.calls[0]).toContain(error);
	});

	test('keeps a missing page quiet and plain', () => {
		const shaped = shapeError({
			error: new Error('Not found: /nope'),
			status: 404,
			message: 'Not Found',
		});

		expect(shaped).toEqual({ message: 'Page not found' });
		expect(logged).not.toHaveBeenCalled();
	});
});
