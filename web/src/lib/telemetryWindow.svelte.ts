import type { InsightsWindow } from './dashboardApi.js';
import { patchSettings } from './settingsSync.js';

// Matches the server's own default (internal/settings DefaultTelemetryWindow)
// -- two constants in two languages, so a change to one wants the other.
export const DEFAULT_TELEMETRY_WINDOW: InsightsWindow = '7d';

const STORAGE_KEY = 'telemetryWindow';

function isWindow(value: unknown): value is InsightsWindow {
	return value === '24h' || value === '7d' || value === '30d';
}

// localStorage is a paint-only cache (persist-account-settings/design.md),
// populated after every successful sync with the server -- same shape as
// forgeFilter.svelte.ts. A cached value outside the documented set is
// ignored rather than trusted: this is what the views send as ?window=.
function readCached(): InsightsWindow {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);

		return isWindow(stored) ? stored : DEFAULT_TELEMETRY_WINDOW;
	} catch {
		return DEFAULT_TELEMETRY_WINDOW;
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
// on the overview is the window the root-cause and flaky views open with.
let telemetryWindow = $state<InsightsWindow>(DEFAULT_TELEMETRY_WINDOW);

export function getTelemetryWindow(): InsightsWindow {
	return telemetryWindow;
}

export function setTelemetryWindow(next: InsightsWindow): void {
	telemetryWindow = next;
	writeCache(next);

	void patchSettings({ telemetryWindow: next });
}

// Called once from the root layout with the window the settings fetch
// returned (undefined if it failed, or on a route that doesn't fetch it),
// falling back to the cached value. Reconciles only: it never patches, since
// the server already holds whatever it just handed us.
export function initTelemetryWindow(serverWindow?: InsightsWindow): void {
	telemetryWindow = serverWindow ?? readCached();
	writeCache(telemetryWindow);
}
