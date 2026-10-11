import { signIn } from './auth.js';
import { expect, test } from './fixtures.js';

// Content that arrives after first paint must not push anything the visitor
// is looking at (#584). Lighthouse found two: the passkey rows on Settings
// (0.029) and the footer under the release history's "Loading…" (0.063).
// Both are phone-sized in Lighthouse, so the viewport is too.
test.use({ viewport: { width: 412, height: 823 } });

async function hold(
	page: import('@playwright/test').Page,
	pattern: string,
	body: unknown,
) {
	let release = () => {};
	const held = new Promise<void>((resolve) => {
		release = resolve;
	});
	await page.route(pattern, async (route) => {
		await held;
		await route.fulfill({ json: body });
	});

	return release;
}

test('the passkey rows arriving do not move what is below them', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	const release = await hold(page, '**/api/auth/credentials', [
		{
			id: 'a',
			label: 'Laptop',
			createdAt: '2026-01-02T00:00:00Z',
			revocable: false,
		},
	]);

	await page.goto('/settings');
	const add = page.getByRole('button', { name: 'Add a passkey' });
	await expect(add).toBeVisible();
	const before = (await add.boundingBox())?.y;

	release();
	await expect(page.getByRole('cell', { name: 'Laptop' })).toBeVisible();

	// Within a pixel: the row's border is a half pixel off the placeholder's.
	expect(
		Math.abs(((await add.boundingBox())?.y ?? 0) - (before ?? 0)),
	).toBeLessThan(1);
});

test('the footer is below the fold while the release history loads', async ({
	page,
	context,
}) => {
	await signIn(page, context);
	const release = await hold(page, '**/releases.json', []);

	await page.goto('/releases');
	await expect(page.getByText('Loading')).toBeVisible();

	const top = await page.evaluate(
		() => document.querySelector('footer')?.getBoundingClientRect().top ?? 0,
	);
	expect(top).toBeGreaterThanOrEqual(823);
	release();
});
