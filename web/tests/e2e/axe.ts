import AxeBuilder from '@axe-core/playwright';
import type { Page } from '@playwright/test';
import { expect } from './fixtures.js';

// WCAG 2.1 AA, the tags rules/a11y.md names.
export const a11yTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

// Any violation fails the run: a scan that only reports is a scan nobody
// acts on.
export async function expectNoViolations(page: Page): Promise<void> {
	const results = await new AxeBuilder({ page }).withTags(a11yTags).analyze();

	expect(results.violations).toEqual([]);
}
