// Carries the response status a caller needs to tell "not found" apart from
// any other failure (the pipeline and usage pages both do), without every
// caller re-parsing a generic Error's message to get it back.
export class ApiError extends Error {
	status: number;

	constructor(status: number) {
		super(`Request failed with status ${status}`);
		this.status = status;
	}
}
