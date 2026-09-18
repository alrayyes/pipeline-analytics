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
} from '$lib/components/ui/alert-dialog/index.js';
import { Button } from '$lib/components/ui/button/index.js';
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from '$lib/components/ui/card/index.js';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from '$lib/components/ui/dialog/index.js';
import { Input } from '$lib/components/ui/input/index.js';
import { Label } from '$lib/components/ui/label/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '$lib/components/ui/table/index.js';
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

interface Credential {
	id: string;
	label: string;
	createdAt: string;
}

let credentials = $state<Credential[]>([]);
let credentialsError = $state<string | null>(null);

async function loadCredentials(): Promise<void> {
	try {
		const res = await fetch('/api/auth/credentials');
		if (!res.ok) {
			credentialsError = 'Could not load passkeys.';

			return;
		}

		credentials = await res.json();
	} catch {
		credentialsError = 'Could not reach the server.';
	}
}

onMount(loadCredentials);

let addOpen = $state(false);
let addLabel = $state('');
let addBusy = $state(false);
let addError = $state<string | null>(null);

function describeError(err: unknown, fallback: string): string {
	if (err instanceof DOMException && err.name === 'NotAllowedError') {
		return 'Passkey ceremony was cancelled or timed out.';
	}

	return err instanceof Error ? err.message : fallback;
}

async function handleAddCredential(event: SubmitEvent): Promise<void> {
	event.preventDefault();
	addBusy = true;
	addError = null;

	try {
		const optionsRes = await fetch('/api/auth/credentials/options', {
			method: 'POST',
		});
		if (!optionsRes.ok) {
			throw new Error('Could not start the ceremony.');
		}

		const optionsBody: { publicKey: PublicKeyCredentialCreationOptionsJSON } =
			await optionsRes.json();
		const options = PublicKeyCredential.parseCreationOptionsFromJSON(
			optionsBody.publicKey,
		);
		const credential = await navigator.credentials.create({
			publicKey: options,
		});

		if (!(credential instanceof PublicKeyCredential)) {
			throw new Error('Registration was not completed.');
		}

		const finishRes = await fetch(
			`/api/auth/credentials?label=${encodeURIComponent(addLabel)}`,
			{
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(credential.toJSON()),
			},
		);
		if (!finishRes.ok) {
			throw new Error('Adding the passkey failed.');
		}

		addOpen = false;
		addLabel = '';
		await loadCredentials();
	} catch (err) {
		addError = describeError(err, 'Adding the passkey failed.');
	} finally {
		addBusy = false;
	}
}

let revokeTarget = $state<Credential | null>(null);
let revokeBusy = $state(false);
let revokeError = $state<string | null>(null);

async function handleRevoke(): Promise<void> {
	if (!revokeTarget) return;

	revokeBusy = true;
	revokeError = null;

	try {
		const res = await fetch(`/api/auth/credentials/${revokeTarget.id}`, {
			method: 'DELETE',
		});
		if (!res.ok) {
			revokeError = 'Could not revoke the passkey.';

			return;
		}

		revokeTarget = null;
		await loadCredentials();
	} catch {
		revokeError = 'Could not reach the server.';
	} finally {
		revokeBusy = false;
	}
}
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

	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Passkeys</CardTitle>
		</CardHeader>
		<CardContent>
			{#if credentialsError}
				<p role="alert" class="text-sm text-destructive">{credentialsError}</p>
			{:else}
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>Label</TableHead>
							<TableHead>Added</TableHead>
							<TableHead class="sr-only">Revoke</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#each credentials as credential (credential.id)}
							<TableRow>
								<TableCell class="font-medium">
									{credential.label || '(unlabeled)'}
								</TableCell>
								<TableCell>
									{new Date(credential.createdAt).toLocaleDateString(undefined, {
										year: 'numeric',
										month: 'long',
										day: 'numeric',
									})}
								</TableCell>
								<TableCell>
									<!--
									Revoking the account's last remaining credential is
									rejected server-side (support-multiple-passkeys/design.md's
									"last-credential guard") -- disabled here rather than
									hidden, matching this app's existing convention for a
									currently-inapplicable action (see "Reset filters" on the
									Pipelines page).
									-->
									<Button
										variant="ghost"
										size="sm"
										disabled={credentials.length <= 1}
										onclick={() => (revokeTarget = credential)}
									>
										Revoke
									</Button>
								</TableCell>
							</TableRow>
						{/each}
					</TableBody>
				</Table>
			{/if}

			<Dialog
				bind:open={addOpen}
				onOpenChange={(open) => {
					if (open) {
						addLabel = '';
						addError = null;
					}
				}}
			>
				<DialogTrigger>
					{#snippet child({ props })}
						<Button {...props} class="mt-4">Add a passkey</Button>
					{/snippet}
				</DialogTrigger>
				<DialogContent>
					<form onsubmit={handleAddCredential}>
						<DialogHeader>
							<DialogTitle>Add a passkey</DialogTitle>
							<DialogDescription>
								Registers a new passkey against this account, so it can log in
								from another device too.
							</DialogDescription>
						</DialogHeader>

						<div class="grid gap-4 py-4">
							<div class="grid gap-2">
								<Label for="credential-label">Label</Label>
								<Input
									id="credential-label"
									bind:value={addLabel}
									placeholder="MacBook, iPhone, …"
									required
								/>
							</div>

							{#if addError}
								<p role="alert" class="text-sm text-destructive">{addError}</p>
							{/if}
						</div>

						<DialogFooter>
							<Button type="submit" disabled={addBusy}>
								{addBusy ? 'Adding…' : 'Add passkey'}
							</Button>
						</DialogFooter>
					</form>
				</DialogContent>
			</Dialog>
		</CardContent>
	</Card>
</main>

<AlertDialog
	open={revokeTarget !== null}
	onOpenChange={(open) => !open && (revokeTarget = null)}
>
	<AlertDialogContent>
		<AlertDialogHeader>
			<AlertDialogTitle>
				Revoke {revokeTarget?.label || 'this passkey'}?
			</AlertDialogTitle>
			<AlertDialogDescription>
				This passkey will no longer be able to log in to this account.
			</AlertDialogDescription>
		</AlertDialogHeader>
		{#if revokeError}
			<p role="alert" class="text-sm text-destructive">{revokeError}</p>
		{/if}
		<AlertDialogFooter>
			<AlertDialogCancel disabled={revokeBusy}>Cancel</AlertDialogCancel>
			<AlertDialogAction disabled={revokeBusy} onclick={handleRevoke}>
				{revokeBusy ? 'Revoking…' : 'Revoke'}
			</AlertDialogAction>
		</AlertDialogFooter>
	</AlertDialogContent>
</AlertDialog>
