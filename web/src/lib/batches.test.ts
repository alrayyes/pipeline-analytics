import { describe, expect, test } from 'bun:test';
import { revealInBatches } from './batches.js';

function queue() {
	const tasks: (() => void)[] = [];

	return {
		schedule: (fn: () => void) => {
			tasks.push(fn);
		},
		run: () => tasks.shift()?.(),
		pending: () => tasks.length,
	};
}

describe('revealInBatches', () => {
	test('reveals the first batch at once, then one step per tick', () => {
		const q = queue();
		const seen: number[] = [];

		revealInBatches(12, 5, 5, (n) => seen.push(n), q.schedule);

		expect(seen).toEqual([5]);
		q.run();
		expect(seen).toEqual([5, 10]);
		q.run();
		expect(seen).toEqual([5, 10, 12]);
		expect(q.pending()).toBe(0);
	});

	test('reveals everything at once when it fits in the first batch', () => {
		const q = queue();
		const seen: number[] = [];

		revealInBatches(3, 5, 5, (n) => seen.push(n), q.schedule);

		expect(seen).toEqual([3]);
		expect(q.pending()).toBe(0);
	});

	test('reveals nothing for an empty list', () => {
		const q = queue();
		const seen: number[] = [];

		revealInBatches(0, 5, 5, (n) => seen.push(n), q.schedule);

		expect(seen).toEqual([]);
		expect(q.pending()).toBe(0);
	});

	test('stops when cancelled', () => {
		const q = queue();
		const seen: number[] = [];

		const cancel = revealInBatches(12, 5, 5, (n) => seen.push(n), q.schedule);
		cancel();
		q.run();

		expect(seen).toEqual([5]);
	});
});
