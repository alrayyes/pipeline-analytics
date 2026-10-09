type Schedule = (fn: () => void) => void;

const nextTask: Schedule = (fn) => {
	setTimeout(fn, 0);
};

// Reveals a long list in slices instead of all at once: `first` items
// straight away, then `step` more per task until `total`. Each slice is its
// own task, so a page that renders markdown for every item never holds the
// main thread for one long task (Lighthouse's total-blocking-time, #522).
// Returns a cancel function for when the component goes away.
export function revealInBatches(
	total: number,
	first: number,
	step: number,
	onReveal: (shown: number) => void,
	schedule: Schedule = nextTask,
): () => void {
	let cancelled = false;

	const reveal = (shown: number) => {
		if (cancelled || total === 0) return;
		const next = Math.min(shown, total);
		onReveal(next);
		if (next < total) schedule(() => reveal(next + step));
	};

	reveal(first);

	return () => {
		cancelled = true;
	};
}
