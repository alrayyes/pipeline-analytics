import { signIn } from './auth.js';
import {
	expectFitsAndPassesAxeInBothSchemes,
	expectNoViolations,
} from './axe.js';
import { expect, test } from './fixtures.js';

// What a visitor must never see, from rules/error-pages.md and the ticket
// (#576): a stack trace, a file path, a framework or service name, a trace ID.
const INTERNALS =
	/stack|trace|\.go:\d|\.ts:\d|SyntaxError|Unexpected token|cluster|namespace|ERR_|drivers\/|services\//i;

test('a path no route matches shows a not-found view in the dashboard layout', async ({
	page,
	context,
}) => {
	await signIn(page, context);

	const response = await page.goto('/no-such-page');

	// A real 404, not a soft one (rules/error-pages.md): the server answers
	// with the app shell and the status, and this view renders in it.
	expect(response?.status()).toBe(404);
	await expect(
		page.getByRole('heading', { level: 1, name: 'Page not found' }),
	).toBeVisible();
	await expect(page).toHaveTitle(/^Page not found/);
	// The shared layout: navigation and footer are still there.
	await expect(page.getByRole('navigation').first()).toBeVisible();
	await expect(page.getByRole('contentinfo')).toBeVisible();
	await expect(page.locator('body')).not.toContainText(INTERNALS);

	await page.getByRole('link', { name: 'Back to the overview' }).click();
	await expect(page).toHaveURL('/');
});

test('an error thrown while loading shows a generic view with no internals', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	// The layout's load reads this body as JSON; a body that isn't throws a
	// SyntaxError whose message must never reach the page.
	await page.route('**/api/repos?limit=1', (route) =>
		route.fulfill({ status: 200, body: '<<not json>>' }),
	);
	const logged: string[] = [];
	page.on('console', (msg) => {
		if (msg.type() === 'error') logged.push(msg.text());
	});

	await page.goto('/no-such-page');

	await expect(
		page.getByRole('heading', { level: 1, name: 'Something went wrong' }),
	).toBeVisible();
	await expect(page).toHaveTitle(/^Something went wrong/);
	await expect(
		page.getByRole('link', { name: 'Back to the overview' }),
	).toBeVisible();
	await expect(page.locator('body')).not.toContainText(INTERNALS);
	// handleError logs the detail instead.
	expect(logged.join('\n')).toMatch(/SyntaxError|JSON/);
	await expectNoViolations(page);
});

for (const [width, height] of [
	[1280, 800],
	[375, 800],
]) {
	test(`the not-found view passes axe at ${width}px in light and dark mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await page.setViewportSize({ width, height });

		await page.goto('/no-such-page');
		await expect(
			page.getByRole('heading', { name: 'Page not found' }),
		).toBeVisible();

		await expectFitsAndPassesAxeInBothSchemes(page, async () => {
			await expect(
				page.getByRole('heading', { name: 'Page not found' }),
			).toBeVisible();
		});
	});
}

// The files the server falls back to for requests the app can't render
// itself (#577), and the template SvelteKit uses for errors in `handle` and
// the root layout. Self-contained: no script, no app stylesheet.
for (const [file, headline] of [
	['/404.html', 'Page not found'],
	['/500.html', 'Something went wrong'],
] as const) {
	test(`${file} is a self-contained static page`, async ({ page, request }) => {
		const html = await (await request.get(file)).text();

		expect(html).not.toMatch(/<script/i);
		expect(html).not.toMatch(/rel=["']stylesheet["']/i);
		expect(html).not.toMatch(INTERNALS);

		await page.goto(file);
		await expect(
			page.getByRole('heading', { level: 1, name: headline }),
		).toBeVisible();
		await expect(page).toHaveTitle(new RegExp(`^${headline}`));
		await expect(
			page.getByRole('link', { name: 'Back to the overview' }),
		).toHaveAttribute('href', '/');
		await expectNoViolations(page);
	});
}
