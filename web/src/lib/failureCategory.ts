import type { FailureCategory } from './dashboardApi.js';

// Display names for the API's failure categories. Which category a failure
// belongs to is decided by the server (metrics.CategorizeFailure); this only
// says how a category reads.
const LABELS: Record<FailureCategory, string> = {
	infrastructure: 'Infrastructure',
	code_tests: 'Code and tests',
	network_timeouts: 'Network and timeouts',
	config_secrets: 'Config and secrets',
	uncategorised: 'Uncategorised',
};

export function categoryLabel(category: FailureCategory): string {
	return LABELS[category];
}
