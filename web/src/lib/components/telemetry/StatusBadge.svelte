<script lang="ts">
import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
import CircleDashedIcon from '@lucide/svelte/icons/circle-dashed';
import CircleXIcon from '@lucide/svelte/icons/circle-x';
import FlaskConicalIcon from '@lucide/svelte/icons/flask-conical';
import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
import { Badge, type BadgeVariant } from '$lib/components/ui/badge/index.js';
import type { Tone } from '$lib/statusModel.js';

let { tone, label }: { tone: Tone; label: string } = $props();

const VARIANTS: Record<Tone, BadgeVariant> = {
	neutral: 'secondary',
	pass: 'success',
	fail: 'destructive',
	running: 'running',
	flaky: 'flaky',
};
</script>

<!-- The destructive variant's tinted text is 3.98:1 in light mode, so a failure
gets the solid fill, as the pipelines page does. -->
<Badge
	variant={VARIANTS[tone]}
	class={tone === 'fail' ? 'bg-destructive text-white' : ''}
>
	{#if tone === 'pass'}
		<CircleCheckIcon aria-hidden="true" />
	{:else if tone === 'fail'}
		<CircleXIcon aria-hidden="true" />
	{:else if tone === 'running'}
		<LoaderCircleIcon aria-hidden="true" class="motion-safe:animate-spin" />
	{:else if tone === 'flaky'}
		<FlaskConicalIcon aria-hidden="true" />
	{:else}
		<CircleDashedIcon aria-hidden="true" />
	{/if}
	{label}
</Badge>
