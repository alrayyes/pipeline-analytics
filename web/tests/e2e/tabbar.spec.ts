import AxeBuilder from '@axe-core/playwright';
import type { BrowserContext, Page } from '@playwright/test';
import { expect, test } from './fixtures.js';

const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

// A CDP virtual authenticator stands in for a passkey device, as in
// dashboard.spec.ts, then the first-run registration signs the test in.
async function signIn(page: Page, context: BrowserContext): Promise<void> {
	const cdp = await context.newCDPSession(page);
	await cdp.send('WebAuthn.enable');
	await cdp.send('WebAuthn.addVirtualAuthenticator', {
		options: {
			protocol: 'ctap2',
			transport: 'internal',
			hasResidentKey: true,
			hasUserVerification: true,
			isUserVerified: true,
			automaticPresenceSimulation: true,
		},
	});

	await page.goto('/');
	await page.getByRole('button', { name: 'Register your passkey' }).click();
	await expect(page.getByRole('button', { name: 'Log out' })).toBeVisible();
}

async function mockOneRepo(page: Page): Promise<void> {
	await page.route('**/api/repos*', (route) =>
		route.fulfill({
			json: {
				repos: [{ id: 'r1', forge: 'github', identifier: 'alrayyes/payments' }],
			},
		}),
	);
}

const PHONE = { width: 390, height: 844 };

test('on a phone, a bottom tab bar reaches the main views and marks the current one', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.setViewportSize(PHONE);
	await page.goto('/');

	const tabs = page.getByRole('navigation', { name: 'Main views' });
	await expect(tabs).toBeVisible();

	for (const name of ['Overview', 'Pipelines', 'Runs', 'Failures']) {
		await expect(tabs.getByRole('link', { name, exact: true })).toBeVisible();
	}

	await expect(
		tabs.getByRole('link', { name: 'Overview', exact: true }),
	).toHaveAttribute('aria-current', 'page');

	await tabs.getByRole('link', { name: 'Runs', exact: true }).click();
	await expect(page).toHaveURL('/runs');
	await expect(
		tabs.getByRole('link', { name: 'Runs', exact: true }),
	).toHaveAttribute('aria-current', 'page');
	await expect(
		tabs.getByRole('link', { name: 'Overview', exact: true }),
	).not.toHaveAttribute('aria-current', 'page');
});

test('the tab bar sits at the bottom of the screen and covers nothing: the footer stays clickable', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.setViewportSize(PHONE);
	await page.goto('/');

	const tabs = page.getByRole('navigation', { name: 'Main views' });
	const box = await tabs.boundingBox();
	expect(box).not.toBeNull();
	// Flush with the bottom edge of the viewport.
	expect((box?.y ?? 0) + (box?.height ?? 0)).toBeCloseTo(PHONE.height, 0);

	// Playwright only clicks what it can hit: a footer link hidden behind the
	// fixed bar would time out here.
	await page.getByRole('link', { name: 'Privacy & disclaimer' }).click();
	await expect(page).toHaveURL('/legal');
});

test('on a desktop width there is no tab bar: the top nav does the job', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await mockOneRepo(page);
	await page.setViewportSize({ width: 1280, height: 800 });
	await page.goto('/');

	await expect(
		page.getByRole('navigation', { name: 'Main views' }),
	).toBeHidden();
});

test('with no repository registered, the tabs that need one are disabled like the top nav', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	await page.setViewportSize(PHONE);

	const tabs = page.getByRole('navigation', { name: 'Main views' });

	await expect(
		tabs.getByRole('link', { name: 'Overview', exact: true }),
	).toHaveAttribute('href', '/');

	for (const name of ['Pipelines', 'Runs', 'Failures']) {
		const tab = tabs.getByText(name, { exact: true });
		await expect(tab).toHaveAttribute('aria-disabled', 'true');
		await expect(tab).not.toHaveAttribute('href');
	}
});

test('there is no tab bar on the login page', async ({ page }) => {
	await page.setViewportSize(PHONE);
	await page.goto('/login');

	await expect(
		page.getByRole('navigation', { name: 'Main views' }),
	).toHaveCount(0);
});

for (const scheme of ['light', 'dark'] as const) {
	test(`the tab bar passes axe and fits a phone in ${scheme} mode`, async ({
		page,
		context,
	}) => {
		await signIn(page, context);
		await mockOneRepo(page);
		await page.setViewportSize(PHONE);
		await page.emulateMedia({ colorScheme: scheme });
		await page.goto('/');
		await expect(
			page.getByRole('navigation', { name: 'Main views' }),
		).toBeVisible();

		const overflow = await page.evaluate(
			() => document.documentElement.scrollWidth - window.innerWidth,
		);
		expect(overflow).toBeLessThanOrEqual(0);

		const scan = await new AxeBuilder({ page }).withTags(a11yTags).analyze();
		expect(scan.violations).toEqual([]);
	});
}
