import { expect, test } from './fixtures.js';

// Installability (#77): the manifest and service worker are both public,
// unauthenticated static assets -- a browser has to be able to fetch them
// before anyone's logged in, the same way the login page itself loads.
test('serves a valid manifest and registers an active service worker', async ({
	page,
}) => {
	const manifestResponse = await page.request.get('/manifest.json');
	expect(manifestResponse.ok()).toBe(true);

	const manifest = await manifestResponse.json();
	expect(manifest.name).toBe('pipeline-analytics');
	expect(manifest.display).toBe('standalone');
	expect(manifest.icons.length).toBeGreaterThanOrEqual(2);
	expect(
		manifest.icons.some(
			(icon: { purpose?: string }) => icon.purpose === 'maskable',
		),
	).toBe(true);

	for (const icon of manifest.icons as { src: string }[]) {
		const iconResponse = await page.request.get(icon.src);
		expect(iconResponse.ok()).toBe(true);
	}

	await page.goto('/login');
	await expect(page.locator('link[rel="manifest"]')).toHaveAttribute(
		'href',
		'/manifest.json',
	);

	const registration = await page.evaluate(async () => {
		const reg = await navigator.serviceWorker.ready;

		return reg.active?.state ?? null;
	});
	expect(registration).toBe('activated');
});
