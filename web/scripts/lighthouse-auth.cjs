// lhci's puppeteerScript: signs the browser in before the signed-in pages are
// audited (#520). A Chrome virtual authenticator stands in for a passkey
// device and the first-run registration signs the browser in, as in
// tests/e2e/auth.ts. lhci runs this before every URL in the same browser, and
// the session cookie outlives it (the config turns the storage reset off), so
// only the first call has anything to do.
//
// Nothing here is logged: the session value stays in the browser.

const ORIGIN = 'http://localhost:4191';

/**
 * @param {import('puppeteer-core').Browser} browser
 */
module.exports = async (browser) => {
	const page = await browser.newPage();
	try {
		await page.goto(`${ORIGIN}/`, { waitUntil: 'networkidle0' });
		if (await page.$('::-p-text(Log out)')) {
			return;
		}

		const cdp = await page.createCDPSession();
		await cdp.send('WebAuthn.enable');
		await cdp.send('WebAuthn.addVirtualAuthenticator', {
			options: {
				protocol: 'ctap2',
				transport: 'internal',
				hasResidentKey: true,
				hasUserVerification: true,
				isUserVerified: true,
				automaticPresenceSimulation: true,
			},
		});

		const register = await page.waitForSelector(
			'::-p-text(Register your passkey)',
		);
		await register.click();
		await page.waitForSelector('::-p-text(Log out)');
	} finally {
		await page.close();
	}
};
