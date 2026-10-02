// Turns a run's recorded steps into the segments of a stage progression bar
// and the one-line summary above it. "Stage" means step: forges expose
// workflows, jobs and steps, but no stage taxonomy.

import type { RunStep } from './dashboardApi.js';
import { statusLabel, statusTone, type Tone } from './statusModel.js';

export interface StageSegment {
	name: string;
	tone: Tone;
	label: string;
}

export function stageSegments(steps: RunStep[]): StageSegment[] {
	return steps.map((step) => ({
		name: step.name,
		tone: statusTone(step.status, step.conclusion),
		label: statusLabel(step.status, step.conclusion),
	}));
}

export function stageSummary(steps: RunStep[]): string {
	if (steps.length === 0) return 'No steps recorded';

	const tones = stageSegments(steps).map((segment) => segment.tone);
	// A failure outranks a stage still running after it: that's the one
	// worth naming.
	const focus = [tones.indexOf('fail'), tones.indexOf('running')].find(
		(index) => index >= 0,
	);

	if (focus !== undefined) {
		return `Stage ${focus + 1}/${steps.length}: ${steps[focus].name}`;
	}

	const passed = tones.filter((tone) => tone === 'pass').length;

	return `${passed}/${steps.length} passed`;
}
