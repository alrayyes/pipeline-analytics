import { describe, expect, test } from 'bun:test';
import { formatPercentagePoints } from './format.js';

describe('formatPercentagePoints', () => {
	test('a drop reads with a minus sign and one decimal', () => {
		expect(formatPercentagePoints(-4.2)).toBe('-4.2 pts');
	});

	test('a rise reads with a plus sign', () => {
		expect(formatPercentagePoints(1.5)).toBe('+1.5 pts');
	});

	test('no change reads as zero, signless', () => {
		expect(formatPercentagePoints(0)).toBe('0.0 pts');
	});

	test('a change that rounds to nothing reads as zero, not +0.0 or -0.0', () => {
		expect(formatPercentagePoints(0.04)).toBe('0.0 pts');
		expect(formatPercentagePoints(-0.04)).toBe('0.0 pts');
	});

	test('rounds to one decimal', () => {
		expect(formatPercentagePoints(-4.25)).toBe('-4.3 pts');
	});
});
