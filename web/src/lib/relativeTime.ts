const UNITS: { unit: Intl.RelativeTimeFormatUnit; seconds: number }[] = [
	{ unit: 'year', seconds: 31536000 },
	{ unit: 'month', seconds: 2592000 },
	{ unit: 'week', seconds: 604800 },
	{ unit: 'day', seconds: 86400 },
	{ unit: 'hour', seconds: 3600 },
	{ unit: 'minute', seconds: 60 },
];

const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });

// Coarsest unit with at least 1 whole tick past the second's worth --
// "5 minutes ago" reads better on a monitoring dashboard than "312 seconds
// ago", and nothing here needs second-level precision.
export function formatRelativeTime(date: Date, now: Date = new Date()): string {
	const diffSeconds = (date.getTime() - now.getTime()) / 1000;

	for (const { unit, seconds } of UNITS) {
		if (Math.abs(diffSeconds) >= seconds) {
			return formatter.format(Math.round(diffSeconds / seconds), unit);
		}
	}

	return formatter.format(0, 'minute');
}
