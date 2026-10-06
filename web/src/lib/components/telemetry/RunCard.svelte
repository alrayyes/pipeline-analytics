<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
import XIcon from '@lucide/svelte/icons/x';
import StageProgress from '#lib/components/telemetry/StageProgress.svelte';
import StatusBadge from '#lib/components/telemetry/StatusBadge.svelte';
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from '#lib/components/ui/alert-dialog/index.js';
import { Button } from '#lib/components/ui/button/index.js';
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from '#lib/components/ui/card/index.js';
import type { RunSummary } from '#lib/dashboardApi.js';
import { formatSeconds } from '#lib/format.js';
import { formatRelativeTime } from '#lib/relativeTime.js';
import {
	actionMessage,
	cancelRun,
	offersCancel,
	offersRerun,
	type RunActionKind,
	rerunRun,
} from '#lib/runActions.js';
import { outcomeLabel, outcomeTone } from '#lib/statusModel.js';

let { run }: { run: RunSummary } = $props();

let busy = $state(false);
let confirmingCancel = $state(false);
let note = $state<{ text: string; failed: boolean } | null>(null);

const ACCEPTED: Record<RunActionKind, string> = {
	rerun: 'Re-run requested. The forge will start it shortly.',
	cancel: 'Cancel requested. The forge will stop it shortly.',
};

async function act(kind: RunActionKind): Promise<void> {
	busy = true;
	note = null;

	const result = await (kind === 'rerun' ? rerunRun : cancelRun)(run.id);
	note = result.ok
		? { text: ACCEPTED[kind], failed: false }
		: { text: actionMessage(kind, result), failed: true };
	busy = false;
}

async function confirmCancel(): Promise<void> {
	confirmingCancel = false;
	await act('cancel');
}

const hasCommit = $derived(Boolean(run.sha || run.message || run.actor));
</script>

<Card class="min-w-0">
	<CardHeader class="flex flex-row items-center justify-between gap-2">
		<CardTitle class="contents">
			<h2 class="min-w-0 truncate">{run.pipelineName}</h2>
		</CardTitle>
		<StatusBadge
			tone={outcomeTone(run.outcome)}
			label={outcomeLabel(run.outcome, run.conclusion ?? run.status)}
		/>
	</CardHeader>
	<CardContent class="grid min-w-0 gap-3">
		<StageProgress steps={run.steps} />
		{#if hasCommit}
			<div class="flex min-w-0 items-center gap-2 text-sm">
				{#if run.sha}
					<code class="shrink-0 font-mono text-xs">{run.sha.slice(0, 7)}</code>
				{/if}
				{#if run.message}
					<span class="min-w-0 truncate">{run.message}</span>
				{/if}
				{#if run.actor}
					<span class="shrink-0 text-muted-foreground">@{run.actor}</span>
				{/if}
			</div>
		{/if}
	</CardContent>
	<CardFooter class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
		{#if run.durationSeconds !== undefined}
			<span>{formatSeconds(run.durationSeconds)}</span>
		{/if}
		{#if run.startedAt}
			<span>{formatRelativeTime(new Date(run.startedAt), new Date())}</span>
		{/if}
		{#if run.forgeUrl}
			<a
				href={run.forgeUrl}
				target="_blank"
				rel="noopener noreferrer"
				class="inline-flex items-center gap-1 text-foreground underline underline-offset-4"
			>
				View run on forge
				<ExternalLinkIcon aria-hidden="true" class="size-3" />
			</a>
		{/if}
		{#if offersRerun(run)}
			<Button
				variant="outline"
				size="sm"
				class="ml-auto"
				disabled={busy}
				onclick={() => act('rerun')}
			>
				<RotateCcwIcon aria-hidden="true" class="size-3" />
				Re-run
			</Button>
		{/if}
		{#if offersCancel(run)}
			<Button
				variant="outline"
				size="sm"
				class="ml-auto"
				disabled={busy}
				onclick={() => (confirmingCancel = true)}
			>
				<XIcon aria-hidden="true" class="size-3" />
				Cancel run
			</Button>
		{/if}
		<p
			role={note?.failed ? 'alert' : 'status'}
			class={['w-full', note?.failed && 'text-destructive']}
		>
			{note?.text ?? ''}
		</p>
	</CardFooter>
</Card>

<AlertDialog bind:open={confirmingCancel}>
	<AlertDialogContent>
		<AlertDialogHeader>
			<AlertDialogTitle>Cancel {run.pipelineName}?</AlertDialogTitle>
			<AlertDialogDescription>
				The forge stops the run and its jobs. A cancelled run can be re-run
				afterwards.
			</AlertDialogDescription>
		</AlertDialogHeader>
		<AlertDialogFooter>
			<AlertDialogCancel>Keep running</AlertDialogCancel>
			<AlertDialogAction onclick={confirmCancel}>Cancel run</AlertDialogAction>
		</AlertDialogFooter>
	</AlertDialogContent>
</AlertDialog>
