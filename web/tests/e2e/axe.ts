import AxeBuilder from '@axe-core/playwright';
import type { Page } from '@playwright/test';
import { expect } from './fixtures.js';

// WCAG 2.1 AA, the tags rules/a11y.md names.
export const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

// Axe reads computed colours, so a button part-way through a colour
// transition reads as a colour that is never on screen (#570). Wait for the
// transitions and animations that end before scanning. Looping ones, like a
// spinner, are left running.
export async function scanSettled(page: Page) {
	await page.evaluate(() =>
		Promise.all(
			document
				.getAnimations()
				// A spinner never finishes; waiting on it would hang the scan.
				.filter((a) => a.effect?.getComputedTiming().iterations !== Infinity)
				// finished rejects with an AbortError when the animation is
				// cancelled mid-wait; a cancelled animation is settled too.
				.map((a) => a.finished.catch(() => undefined)),
		),
	);

	return new AxeBuilder({ page }).withTags(a11yTags).analyze();
}

// Any violation fails the run: a scan that only reports is a scan nobody
// acts on.
//
// The app renders client-side, so a page has no <title> until its route has
// mounted; scanning earlier fails axe's document-title rule (#536). Wait for
// the title first.
export async function expectNoViolations(page: Page): Promise<void> {
	await expect(page).toHaveTitle(/.+/, { timeout: 15_000 });

	const results = await scanSettled(page);

	expect(results.violations).toEqual([]);
}
