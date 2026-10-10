<script lang="ts">
import ShieldIcon from '@lucide/svelte/icons/shield';
import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
import { Button } from '#lib/components/ui/button/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '#lib/components/ui/card/index.js';
import { Input } from '#lib/components/ui/input/index.js';
import { Label } from '#lib/components/ui/label/index.js';
import {
	type FlakyStep,
	type QuarantineState,
	quarantineStep,
	unquarantineStep,
} from '#lib/dashboardApi.js';
import { flakyRunsHref } from '#lib/flakyRuns.js';
import { formatRelativeTime } from '#lib/relativeTime.js';
import FlakeMatrix from './FlakeMatrix.svelte';
import StatusBadge from './StatusBadge.svelte';

// The server caps a note at 500 characters (422 beyond that).
const NOTE_MAX = 500;

let {
	step,
	onChange,
}: { step: FlakyStep; onChange: (state: QuarantineState) => void } = $props();

let asking = $state(false);
let note = $state('');
let busy = $state(false);
let failed = $state(false);

const noteId = $props.id();

async function apply(
	request: () => Promise<QuarantineState>,
): Promise<boolean> {
	busy = true;
	failed = false;

	try {
		onChange(await request());

		return true;
	} catch {
		failed = true;

		return false;
	} finally {
		busy = false;
	}
}

async function quarantine(event: SubmitEvent): Promise<void> {
	event.preventDefault();

	if (
		await apply(() =>
			quarantineStep(step.pipelineId, step.name, note.trim() || undefined),
		)
	) {
		asking = false;
		note = '';
	}
}
</script>

<Card class="min-w-0">
	<CardHeader class="flex flex-row flex-wrap items-center justify-between gap-2">
		<CardTitle class="contents">
			<h3 class="min-w-0 break-words">{step.name}</h3>
		</CardTitle>
		<div class="flex flex-wrap items-center gap-2">
			{#if step.quarantined}
				<span
					class="inline-flex items-center gap-1 rounded-md border border-foreground px-2 py-0.5 text-xs font-medium"
				>
					<ShieldCheckIcon aria-hidden="true" class="size-3" />
					Quarantined
				</span>
			{/if}
			<StatusBadge tone="flaky" label="Flaky" />
		</div>
	</CardHeader>
	<CardContent class="grid min-w-0 gap-3 text-sm">
		<p class="text-muted-foreground">{step.pipelineName}</p>
		<FlakeMatrix {step} />

		{#if step.quarantined && step.quarantine}
			<div class="grid gap-1 rounded-md border p-3">
				<p>
					Quarantined
					<time datetime={step.quarantine.quarantinedAt}>
						{formatRelativeTime(new Date(step.quarantine.quarantinedAt))}</time>.
					Expires
					<time datetime={step.quarantine.expiresAt}>
						{formatRelativeTime(new Date(step.quarantine.expiresAt))}</time>.
				</p>
				{#if step.quarantine.note}
					<p class="break-words text-muted-foreground">
						Note: {step.quarantine.note}
					</p>
				{/if}
			</div>
		{/if}

		<div class="flex flex-wrap items-center gap-3">
			{#if step.quarantined}
				<Button
					variant="outline"
					size="sm"
					disabled={busy}
					onclick={() => apply(() => unquarantineStep(step.pipelineId, step.name))}
				>
					<ShieldIcon aria-hidden="true" />
					Un-quarantine
				</Button>
			{:else if !asking}
				<Button
					variant="outline"
					size="sm"
					onclick={() => {
						asking = true;
					}}
				>
					<ShieldCheckIcon aria-hidden="true" />
					Quarantine
				</Button>
			{/if}
			<a
				href={flakyRunsHref(step.pipelineId, step.name)}
				class="w-fit text-foreground underline underline-offset-4"
			>
				View failed runs
			</a>
		</div>

		{#if asking && !step.quarantined}
			<form class="grid gap-2" onsubmit={quarantine}>
				<Label for={noteId}>Note (optional)</Label>
				<Input id={noteId} bind:value={note} maxlength={NOTE_MAX} autocomplete="off" />
				<div class="flex gap-2">
					<Button type="submit" size="sm" disabled={busy}>Quarantine this step</Button>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onclick={() => {
							asking = false;
							note = '';
							failed = false;
						}}
					>
						Cancel
					</Button>
				</div>
			</form>
		{/if}

		{#if failed}
			<p role="alert" class="text-destructive">
				Couldn't update the quarantine right now.
			</p>
		{/if}
	</CardContent>
</Card>
