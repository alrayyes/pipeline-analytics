<script lang="ts">
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';
import {
	ToggleGroup,
	ToggleGroupItem,
} from '$lib/components/ui/toggle-group/index.js';
import { getTheme, setTheme, type Theme } from '$lib/theme.svelte.js';

const THEME_OPTIONS: { value: Theme; label: string }[] = [
	{ value: 'light', label: 'Light' },
	{ value: 'dark', label: 'Dark' },
	{ value: 'system', label: 'System' },
];
</script>

<svelte:head>
	<title>Settings · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<h1 class="text-2xl font-semibold">Settings</h1>

	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Theme</CardTitle>
		</CardHeader>
		<CardContent>
			<!--
			bits-ui's ToggleGroup root always renders role="group" (its own
			computed props win the merge over anything passed in), but
			type="single" makes each item role="radio" -- which needs a
			radiogroup ancestor per ARIA, or axe's aria-required-parent flags
			it. Supplying that from an outer element here, same pattern as
			ForgeFilter.svelte.
			-->
			<div role="radiogroup" aria-label="Theme">
				<ToggleGroup
					type="single"
					variant="outline"
					value={getTheme()}
					onValueChange={(value) => {
						if (value) setTheme(value as Theme);
					}}
				>
					{#each THEME_OPTIONS as option (option.value)}
						<ToggleGroupItem value={option.value} aria-label={option.label}>
							{option.label}
						</ToggleGroupItem>
					{/each}
				</ToggleGroup>
			</div>
		</CardContent>
	</Card>
</main>
