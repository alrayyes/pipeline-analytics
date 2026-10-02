import { describe, expect, test } from 'bun:test';
import { categoryLabel } from './failureCategory.js';

describe('categoryLabel', () => {
	test.each([
		['infrastructure', 'Infrastructure'],
		['code_tests', 'Code and tests'],
		['network_timeouts', 'Network and timeouts'],
		['config_secrets', 'Config and secrets'],
		['uncategorised', 'Uncategorised'],
	] as const)('%s reads %s', (category, label) => {
		expect(categoryLabel(category)).toBe(label);
	});
});
