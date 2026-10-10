import { signIn } from './auth.js';
import { expectNoViolations } from './axe.js';
import { expect, test } from './fixtures.js';

// The footer renders on every page, the login page included, so the build
// variants can be checked without signing in. The e2e binary is built
// without a release version, which is a dev build; a release build is
// simulated by answering /api/version.

test('a dev build says so and links nowhere', async ({ page }) => {
	await page.goto('/login');

	const footer = page.getByRole('contentinfo');
	await expect(footer.getByText('dev build')).toBeVisible();
	await expect(footer.getByRole('link', { name: /dev/ })).toHaveCount(0);
	await expect(
		footer.getByRole('link', { name: 'Release history' }),
	).toHaveCount(0);
});

test('a release version links to the release history, with no second link to it', async ({
	page,
}) => {
	await page.route('**/api/version', (route) =>
		route.fulfill({ json: { version: '0.56.1' } }),
	);
	await page.goto('/login');

	const footer = page.getByRole('contentinfo');
	await expect(
		footer.getByRole('link', { name: '0.56.1', exact: true }),
	).toHaveAttribute('href', '/releases');
	await expect(
		footer.getByRole('link', { name: 'Release history' }),
	).toHaveCount(0);
	// The privacy link beside it is untouched.
	await expect(
		footer.getByRole('link', { name: 'Privacy & disclaimer' }),
	).toBeVisible();
});

test('the footer links to the GitHub repo and the license (#516)', async ({
	page,
}) => {
	await page.goto('/login');

	const footer = page.getByRole('contentinfo');
	await expect(footer.getByRole('link', { name: 'GitHub' })).toHaveAttribute(
		'href',
		'https://github.com/alrayyes/pipeline-analytics',
	);
	await expect(footer.getByRole('link', { name: 'AGPL-3.0' })).toHaveAttribute(
		'href',
		'https://github.com/alrayyes/pipeline-analytics/blob/main/LICENSE',
	);
	// The mark sits beside its own label, so it is decorative.
	await expect(footer.locator('svg[aria-hidden="true"]')).toHaveCount(1);
});

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

// The two links that leave the app open in a new tab, so the dashboard stays
// where it was (#554). The app's own pages stay in the same tab.
test('the GitHub and license links open in a new tab, and say so', async ({
	page,
	context,
}) => {
	await page.route('**/api/version', (route) =>
		route.fulfill({ json: { version: '0.56.1' } }),
	);
	await page.goto('/login');

	const footer = page.getByRole('contentinfo');
	for (const name of ['GitHub', 'AGPL-3.0']) {
		const link = footer.getByRole('link', { name });
		await expect(link).toHaveAttribute('target', '_blank');
		await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
		// A screen reader hears it before following the link.
		await expect(link).toHaveAccessibleName(
			new RegExp(`${name}.*opens in a new tab`),
		);
	}

	// Same-tab links keep their name and have no target.
	for (const name of ['0.56.1', 'Privacy & disclaimer']) {
		const link = footer.getByRole('link', { name, exact: true });
		await expect(link).not.toHaveAttribute('target', /.*/);
	}

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
