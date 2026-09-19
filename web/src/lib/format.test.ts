import { describe, expect, test } from 'bun:test';
import { formatRate, formatSeconds } from './format.js';

describe('formatSeconds', () => {
	test('under a minute renders whole seconds', () => {
		expect(formatSeconds(45)).toBe('45s');
	});

	test('a minute or more renders one decimal of minutes', () => {
		expect(formatSeconds(90)).toBe('1.5m');
	});

	test('exactly 60 seconds is the minutes boundary', () => {
		expect(formatSeconds(60)).toBe('1.0m');
	});
});

describe('formatRate', () => {
	test('renders a 0-1 rate as a rounded percentage', () => {
		expect(formatRate(0.3)).toBe('30%');
	});

	test('rounds to the nearest whole percent', () => {
		expect(formatRate(0.005)).toBe('1%');
	});

	test('zero renders as 0%', () => {
		expect(formatRate(0)).toBe('0%');
	});
});
