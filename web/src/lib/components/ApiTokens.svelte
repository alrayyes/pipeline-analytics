<script lang="ts">
import {
	createApiToken,
	DEFAULT_TTL_PRESET,
	expiryFor,
	type IssuedApiToken,
	TTL_PRESETS,
	type TtlPreset,
} from '#lib/apiTokens.js';
import { Button } from '#lib/components/ui/button/index.js';
import {
	ToggleGroup,
	ToggleGroupItem,
} from '#lib/components/ui/toggle-group/index.js';

let preset = $state<TtlPreset>(DEFAULT_TTL_PRESET);
let busy = $state(false);
let error = $state<string | null>(null);
// Held in memory only, so it is gone on reload or navigation: the server
// never returns the raw value again.
let issued = $state<IssuedApiToken | null>(null);

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, {
		year: 'numeric',
		month: 'long',
		day: 'numeric',
	});

async function handleCreate(event: SubmitEvent): Promise<void> {
	event.preventDefault();
	busy = true;
	error = null;

	try {
		issued = await createApiToken(preset);
	} catch (err) {
		error = err instanceof Error ? err.message : 'Could not create the token.';
	} finally {
		busy = false;
	}
}
</script>

<form onsubmit={handleCreate} class="grid gap-4">
	<p class="text-sm text-muted-foreground">
		For scripts and CI. Every token expires; pick how long this one lasts.
	</p>

	<!-- radiogroup wrapper for the same reason as ForgeFilter.svelte. -->
	<div role="radiogroup" aria-label="Token lifetime">
		<ToggleGroup
			type="single"
			variant="outline"
			value={preset}
			onValueChange={(value) => {
				if (value) preset = value as TtlPreset;
			}}
		>
			{#each TTL_PRESETS as option (option.value)}
				<ToggleGroupItem value={option.value} aria-label={option.label}>
					{option.label}
				</ToggleGroupItem>
			{/each}
		</ToggleGroup>
	</div>

	<p class="text-sm" data-testid="token-expiry-preview">
		Expires on {formatDate(expiryFor(preset))}
	</p>

	{#if error}
		<p role="alert" class="text-sm text-destructive">{error}</p>
	{/if}

	<div>
		<Button type="submit" disabled={busy}>
			{busy ? 'Creating…' : 'Create token'}
		</Button>
	</div>
</form>

{#if issued}
	<div class="mt-4 grid gap-2" role="status">
		<p class="text-sm">
			Copy this token now. It is shown once and can't be retrieved later.
		</p>
		<code class="break-all rounded bg-muted p-2 text-sm">{issued.token}</code>
		<p class="text-sm text-muted-foreground">
			Expires on {formatDate(new Date(issued.expiresAt))}
		</p>
	</div>
{/if}
