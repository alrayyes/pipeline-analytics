import { scanSettled } from './axe.js';
import { expect, test } from './fixtures.js';

// A transition that is cancelled while scanSettled waits on it (a dialog
// closing, an element unmounting) rejects Animation.finished with an
// AbortError. A cancelled animation is settled, so the scan must go on (#573).
test('scanning carries on when an animation is cancelled mid-wait', async ({
	page,
}) => {
	await page.setContent(
		'<!doctype html><html lang="en"><head><title>t</title></head><body><main><p id="p" style="margin:0">hi</p></main></body></html>',
	);
	await page.evaluate(() => {
		document
			.getElementById('p')
			?.animate([{ opacity: 0 }, { opacity: 1 }], { duration: 60_000 });
		setTimeout(() => {
			for (const a of document.getAnimations()) a.cancel();
		}, 200);
	});

	const results = await scanSettled(page);

	expect(results.violations).toEqual([]);
});
