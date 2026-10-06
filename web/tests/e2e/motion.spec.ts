import type { BrowserContext, Page } from '@playwright/test';
import { signIn } from './auth.js';
import { expect, test } from './fixtures.js';

// The longest transition or animation any element on the page declares, in
// seconds. Read from computed styles, so it sees what tw-animate-css and the
// components really apply rather than what a stylesheet says.
async function longestMotion(page: Page): Promise<number> {
	return page.evaluate(() => {
		const seconds = (list: string): number[] =>
			list.split(',').map((part) => Number.parseFloat(part) || 0);

		return Math.max(
			0,
			...Array.from(document.querySelectorAll('*')).flatMap((el) => {
				const style = getComputedStyle(el);

				return [
					...seconds(style.transitionDuration),
					...seconds(style.animationDuration),
				];
			}),
		);
	});
}

// A dialog is where tw-animate-css actually applies an animation, and it sits
// behind the session.
async function openRegisterDialog(page: Page, context: BrowserContext) {
	await signIn(page, context);
	await page.goto('/repos');
	await page.getByRole('button', { name: 'Register repository' }).click();
	await expect(page.getByRole('dialog')).toBeVisible();
}

test('with no preference the dialog animates, so the check below can fail', async ({
	page,
	context,
}) => {
	await page.emulateMedia({ reducedMotion: 'no-preference' });
	await openRegisterDialog(page, context);

	expect(await longestMotion(page)).toBeGreaterThan(0.05);
});

test('a reader who asked for reduced motion gets none', async ({
	page,
	context,
}) => {
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await openRegisterDialog(page, context);

	expect(await longestMotion(page)).toBeLessThan(0.01);
});
