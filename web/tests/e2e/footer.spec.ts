import { signIn } from './auth.js';
import { expectNoViolations } from './axe.js';
import { expect, test } from './fixtures.js';

// The footer renders on every page, the login page included, so the build
// variants can be checked without signing in. The e2e binary is built
// without a release version, which is a dev build; a release build is
// simulated by answering /api/version.

test('the footer wraps on a narrow phone without scrolling sideways (#516)', async ({
	page,
}) => {
	await page.setViewportSize({ width: 360, height: 800 });
	await page.route('**/api/version', (route) =>
		route.fulfill({ json: { version: '0.56.1' } }),
	);
	await page.goto('/login');

	await expect(page.getByRole('contentinfo')).toBeVisible();
	const overflows = await page.evaluate(
		() =>
			document.documentElement.scrollWidth >
			document.documentElement.clientWidth,
	);
	expect(overflows).toBe(false);
	// Separators sit between links, so none can hang at the end of a row.
	await expect(page.getByRole('contentinfo')).not.toContainText('·');
});

test('the login page passes axe', async ({ page }) => {
	await page.goto('/login');

	await expectNoViolations(page);
});

// Everything but the login page sits behind the session, so the privacy and
// release pages are reached signed in.
test('the privacy and release pages pass axe', async ({ page, context }) => {
	await page.route('**/api/version', (route) =>
		route.fulfill({ json: { version: '0.56.1' } }),
	);
	await signIn(page, context);

	await page.getByRole('link', { name: 'Privacy & disclaimer' }).click();
	await expect(page).toHaveURL('/legal');
	await expectNoViolations(page);

	await page.getByRole('link', { name: '0.56.1', exact: true }).click();
	await expect(page).toHaveURL('/releases');
	await expectNoViolations(page);
});

// The links that leave the app open in a new tab, so the dashboard stays where
// it was (#554). Their attributes are asserted in Footer.dom.test.ts; opening a
// real tab needs a browser.
test('the GitHub link opens the repo in a new tab and leaves the dashboard where it was', async ({
	page,
	context,
}) => {
	await page.route('**/api/version', (route) =>
		route.fulfill({ json: { version: '0.56.1' } }),
	);
	await page.goto('/login');

	const footer = page.getByRole('contentinfo');

	await context.route('https://github.com/**', (route) =>
		route.fulfill({ body: '<title>repo</title>', contentType: 'text/html' }),
	);
	const [opened] = await Promise.all([
		context.waitForEvent('page'),
		footer.getByRole('link', { name: 'GitHub' }).click(),
	]);
	await opened.waitForLoadState();
	expect(opened.url()).toBe('https://github.com/alrayyes/pipeline-analytics');
	expect(page.url()).toContain('/login');

	await expectNoViolations(page);
});
