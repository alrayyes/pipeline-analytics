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
