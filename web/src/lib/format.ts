// Shared duration/percentage formatting for the Steps tables (pipeline
// detail page and the cross-pipeline Unhealthy steps page).

export function formatSeconds(seconds: number): string {
	return seconds >= 60
		? `${(seconds / 60).toFixed(1)}m`
		: `${seconds.toFixed(0)}s`;
}

export function formatRate(rate: number): string {
	return `${Math.round(rate * 100)}%`;
}
