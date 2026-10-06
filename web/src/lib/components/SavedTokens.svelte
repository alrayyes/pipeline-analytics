<script lang="ts">
import { onMount } from 'svelte';
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
import { Input } from '#lib/components/ui/input/index.js';
import { Label } from '#lib/components/ui/label/index.js';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
} from '#lib/components/ui/select/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '#lib/components/ui/table/index.js';
import {
	deleteForgeToken,
	listForgeTokens,
	type SavedForgeToken,
	savedTokenLabel,
	saveForgeToken,
} from '#lib/savedForgeTokens.js';

const FORGE_LABELS = { github: 'GitHub', forgejo: 'Forgejo' } as const;

let tokens = $state<SavedForgeToken[]>([]);
let loadError = $state<string | null>(null);

async function load(): Promise<void> {
	try {
		tokens = await listForgeTokens();
		loadError = null;
	} catch {
		loadError = 'Could not load saved tokens.';
	}
}

onMount(load);

let forge = $state<'github' | 'forgejo'>('github');
let instanceUrl = $state('');
let token = $state('');
let saveBusy = $state(false);
let saveError = $state<string | null>(null);
let savedNotice = $state<string | null>(null);

async function handleSave(event: SubmitEvent): Promise<void> {
	event.preventDefault();
	saveBusy = true;
	saveError = null;
	savedNotice = null;

	try {
		const saved = await saveForgeToken({
			forge,
			forgejoInstanceUrl: instanceUrl,
			token,
		});

		// The token never leaves the browser's memory longer than this call.
		token = '';
		savedNotice = `Saved ${savedTokenLabel(saved)} token (${saved.tokenMasked}).`;
		await load();
	} catch (err) {
		saveError =
			err instanceof Error ? err.message : 'Could not save the token.';
	} finally {
		saveBusy = false;
	}
}

let deleteTarget = $state<SavedForgeToken | null>(null);
let deleteBusy = $state(false);
let deleteError = $state<string | null>(null);

async function handleDelete(): Promise<void> {
	if (!deleteTarget) return;

	deleteBusy = true;
	deleteError = null;

	try {
		await deleteForgeToken(deleteTarget.id);
		deleteTarget = null;
		await load();
	} catch (err) {
		deleteError =
			err instanceof Error ? err.message : 'Could not delete the token.';
	} finally {
		deleteBusy = false;
	}
}
</script>

<p class="text-sm text-muted-foreground">
	Used when you register a repository, so you don't paste a token again. They
	are stored encrypted on the server and only the last four characters are ever
	shown. Repositories already registered keep working if you delete one.
</p>

{#if loadError}
	<p role="alert" class="mt-4 text-sm text-destructive">{loadError}</p>
{:else if tokens.length === 0}
	<p class="mt-4 text-sm text-muted-foreground">No saved tokens.</p>
{:else}
	<Table class="mt-4">
		<TableHeader>
			<TableRow>
				<TableHead>Forge</TableHead>
				<TableHead>Token</TableHead>
				<TableHead class="sr-only">Delete</TableHead>
			</TableRow>
		</TableHeader>
		<TableBody>
			{#each tokens as saved (saved.id)}
				<TableRow>
					<TableCell class="font-medium break-all">{savedTokenLabel(saved)}</TableCell>
					<TableCell><code>{saved.tokenMasked}</code></TableCell>
					<TableCell>
						<Button
							variant="ghost"
							size="sm"
							aria-label="Delete the {savedTokenLabel(saved)} token"
							onclick={() => {
								deleteError = null;
								deleteTarget = saved;
							}}
						>
							Delete
						</Button>
					</TableCell>
				</TableRow>
			{/each}
		</TableBody>
	</Table>
{/if}

<form class="mt-6 grid gap-4" onsubmit={handleSave}>
	<h3 class="text-sm font-medium">Save a token</h3>

	<div class="grid gap-2">
		<Label for="saved-token-forge">Forge</Label>
		<Select type="single" bind:value={forge}>
			<SelectTrigger id="saved-token-forge" class="w-full sm:w-64">
				{FORGE_LABELS[forge]}
			</SelectTrigger>
			<SelectContent>
				<SelectItem value="github" label="GitHub">GitHub</SelectItem>
				<SelectItem value="forgejo" label="Forgejo">Forgejo</SelectItem>
			</SelectContent>
		</Select>
	</div>

	{#if forge === 'forgejo'}
		<div class="grid gap-2">
			<Label for="saved-token-instance">Forgejo instance URL</Label>
			<Input
				id="saved-token-instance"
				type="url"
				bind:value={instanceUrl}
				placeholder="https://forgejo.example.com"
				required
			/>
		</div>
	{/if}

	<div class="grid gap-2">
		<Label for="saved-token-value">Access token</Label>
		<Input
			id="saved-token-value"
			type="password"
			autocomplete="off"
			bind:value={token}
			required
		/>
		<p class="text-xs text-muted-foreground">
			Saving replaces the token already saved for that forge and instance.
		</p>
	</div>

	{#if saveError}
		<p role="alert" class="text-sm text-destructive">{saveError}</p>
	{/if}
	{#if savedNotice}
		<p role="status" class="text-sm text-muted-foreground">{savedNotice}</p>
	{/if}

	<Button type="submit" class="w-fit" disabled={saveBusy}>
		{saveBusy ? 'Saving…' : 'Save token'}
	</Button>
</form>

<AlertDialog
	open={deleteTarget !== null}
	onOpenChange={(open) => !open && (deleteTarget = null)}
>
	<AlertDialogContent>
		<AlertDialogHeader>
			<AlertDialogTitle>
				Delete the {deleteTarget ? savedTokenLabel(deleteTarget) : ''} token?
			</AlertDialogTitle>
			<AlertDialogDescription>
				Registering a repository will ask for a token again. Repositories
				already registered with it keep working.
			</AlertDialogDescription>
		</AlertDialogHeader>
		{#if deleteError}
			<p role="alert" class="text-sm text-destructive">{deleteError}</p>
		{/if}
		<AlertDialogFooter>
			<AlertDialogCancel disabled={deleteBusy}>Cancel</AlertDialogCancel>
			<AlertDialogAction disabled={deleteBusy} onclick={handleDelete}>
				{deleteBusy ? 'Deleting…' : 'Delete'}
			</AlertDialogAction>
		</AlertDialogFooter>
	</AlertDialogContent>
</AlertDialog>
