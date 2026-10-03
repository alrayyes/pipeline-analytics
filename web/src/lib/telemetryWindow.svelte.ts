import type { InsightsWindow } from './dashboardApi.js';
import { patchSettings } from './settingsSync.js';

// The window the telemetry views cover is the account's saved choice, and the
// frontend holds no default of its own: with nothing saved or cached it is
// null, the views send no window, and the server uses its default and reports
// which in the response, which is what the toggle then shows (#376).

const STORAGE_KEY = 'telemetryWindow';

function isWindow(value: unknown): value is InsightsWindow {
	return value === '24h' || value === '7d' || value === '30d';
}

// localStorage is a paint-only cache (persist-account-settings/design.md),
// populated after every successful sync with the server -- same shape as
// forgeFilter.svelte.ts. A cached value outside the documented set is
// ignored rather than trusted: this is what the views send as ?window=.
function readCached(): InsightsWindow | null {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return isWindow(stored) ? stored : null;
	} catch {
		return null;
	}
}

function writeCache(window: InsightsWindow): void {
	try {
		localStorage.setItem(STORAGE_KEY, window);
	} catch {
		// Best effort -- see theme.svelte.ts's identical comment.
	}
}

// One module-level rune shared by every telemetry view, so a window chosen
// on the overview is the window the root-cause view opens with.
let telemetryWindow = $state<InsightsWindow | null>(null);

export function getTelemetryWindow(): InsightsWindow | null {
	return telemetryWindow;
}

export function setTelemetryWindow(next: InsightsWindow): void {
	telemetryWindow = next;
	writeCache(next);

	void patchSettings({ telemetryWindow: next });
}

// Called once from the root layout with the window the settings fetch
// returned (undefined if it failed, or on a route that doesn't fetch it),
// falling back to the cached value, then to nothing. Reconciles only: it
// never patches, since the server already holds whatever it just handed us.
export function initTelemetryWindow(serverWindow?: InsightsWindow): void {
	telemetryWindow = serverWindow ?? readCached();

	if (telemetryWindow) writeCache(telemetryWindow);
}
