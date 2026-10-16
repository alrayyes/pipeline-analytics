// Carries the response status a caller needs to tell "not found" apart from
// any other failure (the pipeline and usage pages both do), without every
// caller re-parsing a generic Error's message to get it back. A server
// refusal also keeps its `code` and `message`: the message is appended to the
// error's own, since an agent reads only that, and the server's wording is
// what says what to fix.
export class ApiError extends Error {
	status: number;
	code?: string;
	serverMessage?: string;

	constructor(status: number, code?: string, serverMessage?: string) {
		super(`Request failed with status ${status}`);
		this.status = status;
		this.code = code;
		this.serverMessage = serverMessage;
		if (serverMessage) this.message = `${this.message}: ${serverMessage}`;
	}
}

// Reads a refusal's JSON body, if it has one, into an ApiError.
export async function refusal(res: Response): Promise<ApiError> {
	const body = (await res.json().catch(() => null)) as {
		code?: string;
		message?: string;
	} | null;

	return new ApiError(res.status, body?.code, body?.message);
}
