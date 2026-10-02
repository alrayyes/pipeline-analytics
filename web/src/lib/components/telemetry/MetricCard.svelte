<script lang="ts">
import type { Component } from 'svelte';
import { Card, CardContent } from '$lib/components/ui/card/index.js';

let {
	label,
	value,
	note,
	icon: Icon,
}: {
	label: string;
	value: string;
	note?: string;
	icon?: Component<{ class?: string; 'aria-hidden'?: 'true' }>;
} = $props();
</script>

<div role="group" aria-label={label} class="min-w-0">
	<Card class="h-full min-w-0">
		<CardContent class="grid gap-1">
			<p class="text-xs text-muted-foreground">{label}</p>
			<p class="text-2xl font-semibold break-words">{value}</p>
			{#if note}
				<p class="flex items-center gap-1 text-sm text-muted-foreground">
					{#if Icon}<Icon class="size-4 shrink-0" aria-hidden="true" />{/if}
					<span>{note}</span>
				</p>
			{/if}
		</CardContent>
	</Card>
</div>
