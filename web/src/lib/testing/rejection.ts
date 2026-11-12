// The error a promise rejects with. A test that expects a rejection gets a
// failure here, not a value it then reads `.message` off by accident, and the
// result is typed `Error` rather than the union of success and failure that
// `.catch((e) => e as Error)` leaves behind.
export async function rejection(promise: Promise<unknown>): Promise<Error> {
	try {
		await promise;
	} catch (e) {
		return e as Error;
	}

	throw new Error('expected the promise to reject');
}
