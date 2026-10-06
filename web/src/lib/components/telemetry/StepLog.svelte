<script lang="ts">
import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
import { onMount } from 'svelte';
import { parseAnsiLines } from '#lib/ansi.js';
import { fetchJobLog, JOB_LOG_LINES, type JobLog } from '#lib/dashboardApi.js';

let {
	runId,
	jobId,
	forgeUrl,
}: { runId: string; jobId: string; forgeUrl?: string } = $props();

let log = $state<JobLog | null>(null);
let failed = $state(false);

const REASONS: Record<string, string> = {
	unsupported: 'This forge has no log API, so the log can only be read there.',
	expired: 'The forge no longer has this log.',
	forbidden: 'The saved token isn’t allowed to read this log.',
	unreachable: 'The forge didn’t answer.',
};

// Each line becomes text segments, drawn through plain interpolation below:
// never {@html}, so markup in a log is shown, not run.
const lines = $derived(log?.available ? parseAnsiLines(log.lines) : []);
const link = $derived(log?.forgeUrl ?? forgeUrl);

onMount(async () => {
	try {
		log = await fetchJobLog(runId, jobId);
	} catch {
		failed = true;
	}
});
</script>

{#if failed}
	<p role="alert" class="text-sm text-destructive">Could not load the log.</p>
{:else if log === null}
	<p class="text-sm text-muted-foreground">Loading log…</p>
{:else if !log.available}
	<p class="text-sm text-muted-foreground">
		{REASONS[log.reason ?? ''] ?? 'No log is available.'}
	</p>
{:else}
	{#if log.truncated}
		<p class="mb-1 text-xs text-muted-foreground">Last {JOB_LOG_LINES} lines.</p>
	{/if}
	<!-- A scrollable region has to be reachable by keyboard to be scrolled by it. -->
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<pre
		tabindex="0"
		aria-label="Step log"
		class="max-h-96 overflow-auto rounded-md p-3 font-mono text-xs leading-5 whitespace-pre-wrap break-words"
		style="background: var(--terminal-bg); color: var(--terminal-fg);"
	>{#each lines as line, i (i)}<span class="block">{#each line as seg, j (j)}<span class={[seg.fg && `ansi-${seg.fg}`, seg.bold && 'font-bold']}>{seg.text}</span>{:else}&nbsp;{/each}</span>{/each}</pre>
{/if}
{#if link}
	<a
		href={link}
		target="_blank"
		rel="noreferrer"
		class="mt-2 inline-flex items-center gap-1 text-sm text-foreground underline underline-offset-4"
	>
		View on forge
		<ExternalLinkIcon aria-hidden="true" class="size-3" />
	</a>
{/if}
