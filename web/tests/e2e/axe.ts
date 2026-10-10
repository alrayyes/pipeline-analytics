import AxeBuilder from '@axe-core/playwright';
import type { Page } from '@playwright/test';
import { expect } from './fixtures.js';

// WCAG 2.1 AA, the tags rules/a11y.md names.
export const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

// Any violation fails the run: a scan that only reports is a scan nobody
// acts on.
//
// The app renders client-side, so a page has no <title> until its route has
// mounted; scanning earlier fails axe's document-title rule (#536). Wait for
// the title first.
export async function expectNoViolations(page: Page): Promise<void> {
	await expect(page).toHaveTitle(/.+/, { timeout: 15_000 });

	const results = await new AxeBuilder({ page }).withTags(a11yTags).analyze();

	expect(results.violations).toEqual([]);
}
