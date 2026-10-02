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

// A change in percentage points as the server reports it (passRateDelta):
// signed, one decimal. A change that rounds to nothing carries no sign, so
// it never reads "+0.0" or "-0.0".
export function formatPercentagePoints(points: number): string {
	// Round the magnitude, so a half rounds away from zero on both sides
	// (Math.round on the signed value would send -4.25 to -4.2 but 4.25 to
	// 4.3).
	const magnitude = Math.abs(points).toFixed(1);
	if (magnitude === '0.0') return '0.0 pts';

	return `${points > 0 ? '+' : '-'}${magnitude} pts`;
}
