<script lang="ts">
import { onMount } from 'svelte';
import { invalidateAll } from '$app/navigation';
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
import { Badge } from '$lib/components/ui/badge/index.js';
import { Button } from '$lib/components/ui/button/index.js';
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
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
} from '$lib/components/ui/select/index.js';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from '$lib/components/ui/table/index.js';
import {
	getRememberedToken,
	maskToken,
	rememberToken,
} from '$lib/rememberedToken.js';

interface Repo {
	id: string;
	forge: 'github' | 'forgejo';
	identifier: string;
	forgejoInstanceUrl?: string;
	tokenMasked: string;
	ingestionStatus: 'pending' | 'active' | 'degraded';
	ingestionStatusReason?: string;
}

const FORGE_LABELS: Record<string, string> = {
	github: 'GitHub',
	forgejo: 'Forgejo',
};
const TOKEN_HELP: Record<string, string> = {
	github:
		'Classic token: the "repo" scope. Fine-grained token: this repository selected, with "Actions" (read-only) and "Webhooks" (read and write) repository permissions.',
	forgejo:
		'A token with "repository" (read and write) and "user" (read) access, created under Settings → Applications → Manage Access Tokens.',
};
const STATUS_VARIANTS: Record<string, 'default' | 'destructive' | 'outline'> = {
	active: 'default',
	pending: 'outline',
	degraded: 'destructive',
};

let repos = $state<Repo[] | null>(null);
let error = $state<string | null>(null);

let registerOpen = $state(false);
let registerBusy = $state(false);

// Two explicit steps (#103): connect a token, then pick which repos to
// follow with it -- rather than one repo per dialog session, tied to a
// single-select identifier field.
let step = $state<'token' | 'select'>('token');

let forge = $state<'github' | 'forgejo'>('github');
let forgejoInstanceUrl = $state('');
let token = $state('');

let discoveredRepos = $state<string[] | null>(null);
let discoverBusy = $state(false);
let discoverError = $state<string | null>(null);

let selected = $state<Set<string>>(new Set());
let manualIdentifier = $state('');

interface RegisterResult {
	identifier: string;
	ok: boolean;
	message?: string;
}

let batchResults = $state<RegisterResult[] | null>(null);

const canDiscover = $derived(
	token.trim() !== '' &&
		(forge === 'github' || forgejoInstanceUrl.trim() !== ''),
);

const allDiscoveredSelected = $derived(
	(discoveredRepos?.length ?? 0) > 0 &&
		(discoveredRepos ?? []).every((id) => selected.has(id)),
);

// Repos already tracked on this exact forge (+ instance, for Forgejo)
// don't belong in the picker at all -- there's nothing to do with one a
// second time.
const trackedIdentifiers = $derived(
	new Set(
		(repos ?? [])
			.filter(
				(r) =>
					r.forge === forge &&
					(forge !== 'forgejo' || r.forgejoInstanceUrl === forgejoInstanceUrl),
			)
			.map((r) => r.identifier),
	),
);

// Manually-added identifiers aren't necessarily in discoveredRepos (the
// whole point of "add another by name" is covering what discovery didn't
// return), so they need their own list to render.
const manualOnlySelected = $derived(
	[...selected].filter((id) => !discoveredRepos?.includes(id)),
);

// A token already used to register a repo on this forge (+ instance, for
// Forgejo) is offered for reuse rather than requiring it be retyped --
// browser-local only (issue #72's design decision), so it doesn't survive a
// cleared browser or carry across devices. A plain $derived here would only
// re-read localStorage when `forge`/`forgejoInstanceUrl` themselves change
// value, which they often don't between one dialog open and the next (the
// default forge is always 'github') -- keying an $effect on `registerOpen`
// too forces a fresh read every time the dialog opens, picking up a token
// a previous registration just remembered.
let rememberedForCurrentForge = $state<string | null>(null);

$effect(() => {
	[forge, forgejoInstanceUrl, registerOpen];
	rememberedForCurrentForge = getRememberedToken(
		forge,
		forge === 'forgejo' ? forgejoInstanceUrl : undefined,
	);
});

// A discovered list only makes sense for the token/forge/instance it was
// fetched for -- invalidate it (and the selection, which was picked from
// it) the moment any of those change underneath it, rather than letting a
// stale picker suggest repos the current token might not even reach.
$effect(() => {
	[forge, token, forgejoInstanceUrl];
	discoveredRepos = null;
	discoverError = null;
	selected = new Set();
	batchResults = null;
});

let untrackTarget = $state<Repo | null>(null);
let untrackBusy = $state(false);
let untrackError = $state<string | null>(null);

async function loadRepos(): Promise<void> {
	try {
		const res = await fetch('/api/repos');
		if (!res.ok) {
			error = 'Could not load repositories.';

			return;
		}

		repos = await res.json();
	} catch {
		error = 'Could not reach the server.';
	}
}

onMount(loadRepos);

function resetForm(): void {
	step = 'token';
	forge = 'github';
	forgejoInstanceUrl = '';
	token = '';
	discoveredRepos = null;
	discoverError = null;
	selected = new Set();
	manualIdentifier = '';
	batchResults = null;
}

async function discoverRepos(): Promise<void> {
	discoverBusy = true;
	discoverError = null;

	try {
		const res = await fetch('/api/repos/discover', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				forge,
				token,
				...(forge === 'forgejo' ? { forgejoInstanceUrl } : {}),
			}),
		});

		if (!res.ok) {
			const body = await res.json().catch(() => null);
			discoverError = body?.message ?? 'Could not list repositories.';
			discoveredRepos = null;

			return;
		}

		const found: string[] = await res.json();
		discoveredRepos = found.filter((id) => !trackedIdentifiers.has(id));
		step = 'select';
	} catch {
		discoverError = 'Could not reach the server.';
	} finally {
		discoverBusy = false;
	}
}

function handleDiscover(event: SubmitEvent): void {
	event.preventDefault();
	discoverRepos();
}

// One click, not two (previously: fill the token, then separately click
// "Find repositories") -- the whole point of remembering a token is not
// having to redo the rest of the ceremony to add one more repo.
function useSavedTokenAndDiscover(): void {
	if (!rememberedForCurrentForge) return;

	token = rememberedForCurrentForge;
	discoverRepos();
}

function toggleSelected(identifier: string): void {
	const next = new Set(selected);

	if (next.has(identifier)) {
		next.delete(identifier);
	} else {
		next.add(identifier);
	}

	selected = next;
}

function toggleSelectAllDiscovered(): void {
	if (!discoveredRepos) return;

	if (allDiscoveredSelected) {
		const next = new Set(selected);
		for (const id of discoveredRepos) next.delete(id);
		selected = next;
	} else {
		selected = new Set([...selected, ...discoveredRepos]);
	}
}

function addManualIdentifier(): void {
	const id = manualIdentifier.trim();
	if (!id) return;

	selected = new Set([...selected, id]);
	manualIdentifier = '';
}

async function handleFollowSelected(event: SubmitEvent): Promise<void> {
	event.preventDefault();
	if (selected.size === 0) return;

	registerBusy = true;

	const results: RegisterResult[] = [];

	// Sequential, not parallel -- registering N repos at once shouldn't
	// hammer the forge API with N concurrent webhook-creation calls.
	for (const identifier of selected) {
		try {
			const res = await fetch('/api/repos', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					forge,
					identifier,
					token,
					...(forge === 'forgejo' ? { forgejoInstanceUrl } : {}),
				}),
			});

			if (res.ok) {
				results.push({ identifier, ok: true });
			} else {
				const body = await res.json().catch(() => null);
				results.push({
					identifier,
					ok: false,
					message: body?.message ?? 'Could not register.',
				});
			}
		} catch {
			results.push({
				identifier,
				ok: false,
				message: 'Could not reach the server.',
			});
		}
	}

	batchResults = results;
	registerBusy = false;

	// A degraded-but-persisted repo (Registrar.degrade -- its webhook
	// creation failed, but the repo itself is still tracked) still counts
	// as a real POST success here, same as the single-repo flow always
	// treated it.
	rememberToken(
		forge,
		forge === 'forgejo' ? forgejoInstanceUrl : undefined,
		token,
	);

	// The layout's own load -- which the nav's hasRepos-gated links and the
	// Pipelines empty state (#71) both read -- only reruns on navigation by
	// default; this registration didn't navigate anywhere.
	await Promise.all([loadRepos(), invalidateAll()]);

	if (results.every((r) => r.ok)) {
		// Not resetForm() here -- it flips `step` back to 'token', which
		// would swap the dialog's content out from under its own closing
		// animation (step 2's markup replaced by step 1's while still
		// fading out). onOpenChange already resets on the next open.
		registerOpen = false;
	} else {
		// Keep only what failed selected, so retrying the batch doesn't
		// re-submit repos that already registered successfully.
		selected = new Set(results.filter((r) => !r.ok).map((r) => r.identifier));
	}
}

async function handleUntrack(): Promise<void> {
	if (!untrackTarget) return;

	untrackBusy = true;
	untrackError = null;

	try {
		const res = await fetch(`/api/repos/${untrackTarget.id}`, {
			method: 'DELETE',
		});
		if (!res.ok) {
			untrackError = 'Could not untrack the repository.';

			return;
		}

		untrackTarget = null;
		await Promise.all([loadRepos(), invalidateAll()]);
	} catch {
		untrackError = 'Could not reach the server.';
	} finally {
		untrackBusy = false;
	}
}
</script>

<svelte:head>
	<title>Repositories · pipeline-analytics</title>
</svelte:head>

<main class="mx-auto max-w-4xl px-4 py-8">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold">Repositories</h1>
		<Dialog
			bind:open={registerOpen}
			onOpenChange={(open) => {
				if (open) resetForm();
			}}
		>
			<DialogTrigger>
				{#snippet child({ props })}
					<Button {...props}>Register repository</Button>
				{/snippet}
			</DialogTrigger>
			<DialogContent>
				{#if step === 'token'}
					<form onsubmit={handleDiscover}>
						<DialogHeader>
							<DialogTitle>Connect a token</DialogTitle>
							<DialogDescription>
								Then pick which of its repositories to follow.
							</DialogDescription>
						</DialogHeader>

						<div class="grid gap-4 py-4">
							<div class="grid gap-2">
								<Label for="forge">Forge</Label>
								<Select type="single" bind:value={forge}>
									<SelectTrigger id="forge" class="w-full">
										{FORGE_LABELS[forge]}
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="github" label="GitHub">GitHub</SelectItem>
										<SelectItem value="forgejo" label="Forgejo">Forgejo</SelectItem>
									</SelectContent>
								</Select>
							</div>

							<div class="grid gap-2">
								<Label for="token">Access token</Label>
								<Input id="token" type="password" bind:value={token} required />
								{#if rememberedForCurrentForge && token !== rememberedForCurrentForge}
									<button
										type="button"
										class="w-fit text-xs text-muted-foreground underline hover:text-foreground"
										onclick={useSavedTokenAndDiscover}
										disabled={discoverBusy}
									>
										{discoverBusy
											? 'Finding…'
											: `Use saved token (${maskToken(rememberedForCurrentForge)})`}
									</button>
								{/if}
								<p class="text-xs text-muted-foreground">{TOKEN_HELP[forge]}</p>
							</div>

							{#if forge === 'forgejo'}
								<div class="grid gap-2">
									<Label for="instance-url">Forgejo instance URL</Label>
									<Input
										id="instance-url"
										type="url"
										bind:value={forgejoInstanceUrl}
										placeholder="https://forgejo.example.com"
										required
									/>
								</div>
							{/if}

							{#if discoverError}
								<p role="alert" class="text-sm text-destructive">{discoverError}</p>
							{/if}
						</div>

						<DialogFooter>
							<Button type="submit" disabled={!canDiscover || discoverBusy}>
								{discoverBusy ? 'Finding…' : 'Find repositories'}
							</Button>
						</DialogFooter>
					</form>
				{:else}
					<form onsubmit={handleFollowSelected}>
						<DialogHeader>
							<DialogTitle>Select repositories to follow</DialogTitle>
							<DialogDescription>
								{FORGE_LABELS[forge]} · token ending {maskToken(token)}
							</DialogDescription>
						</DialogHeader>

						<div class="grid gap-3 py-4">
							{#if discoveredRepos && discoveredRepos.length > 0}
								<div class="flex items-center justify-between">
									<span class="text-sm text-muted-foreground">
										{selected.size} selected
									</span>
									<button
										type="button"
										class="text-xs text-muted-foreground underline hover:text-foreground"
										onclick={toggleSelectAllDiscovered}
									>
										{allDiscoveredSelected ? 'Deselect all' : 'Select all'}
									</button>
								</div>
								<ul class="grid max-h-64 gap-1 overflow-y-auto rounded-md border p-2">
									{#each discoveredRepos as repoId (repoId)}
										<li>
											<label
												class="flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-muted"
											>
												<input
													type="checkbox"
													class="size-4"
													checked={selected.has(repoId)}
													onchange={() => toggleSelected(repoId)}
												/>
												{repoId}
											</label>
										</li>
									{/each}
								</ul>
							{:else}
								<p class="text-sm text-muted-foreground">
									That token can't see any repositories.
								</p>
							{/if}

							{#if manualOnlySelected.length > 0}
								<ul class="grid gap-1">
									{#each manualOnlySelected as id (id)}
										<li
											class="flex items-center justify-between rounded border px-2 py-1.5 text-sm"
										>
											{id}
											<button
												type="button"
												class="text-muted-foreground hover:text-foreground"
												aria-label="Remove {id}"
												onclick={() => toggleSelected(id)}
											>
												×
											</button>
										</li>
									{/each}
								</ul>
							{/if}

							<div class="grid gap-2">
								<Label for="manual-identifier">Add another by name</Label>
								<div class="flex gap-2">
									<Input
										id="manual-identifier"
										bind:value={manualIdentifier}
										placeholder="owner/name"
										class="flex-1"
										onkeydown={(event) => {
											if (event.key !== 'Enter') return;
											event.preventDefault();
											addManualIdentifier();
										}}
									/>
									<Button
										type="button"
										variant="outline"
										disabled={!manualIdentifier.trim()}
										onclick={addManualIdentifier}
									>
										Add
									</Button>
								</div>
							</div>

							{#if batchResults}
								<ul class="grid gap-1 text-sm">
									{#each batchResults as result (result.identifier)}
										<li class={result.ok ? 'text-muted-foreground' : 'text-destructive'}>
											{result.identifier}: {result.ok ? 'registered' : result.message}
										</li>
									{/each}
								</ul>
							{/if}
						</div>

						<DialogFooter class="sm:justify-between">
							<Button type="button" variant="ghost" onclick={() => (step = 'token')}>
								Back
							</Button>
							<Button type="submit" disabled={selected.size === 0 || registerBusy}>
								{registerBusy
									? 'Registering…'
									: `Follow ${selected.size} ${selected.size === 1 ? 'repository' : 'repositories'}`}
							</Button>
						</DialogFooter>
					</form>
				{/if}
			</DialogContent>
		</Dialog>
	</div>

	{#if error}
		<p role="alert" class="mt-6 text-destructive">{error}</p>
	{:else if repos === null}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if repos.length === 0}
		<p class="mt-6 text-muted-foreground">No repositories tracked yet.</p>
	{:else}
		<Table class="mt-6">
			<TableHeader>
				<TableRow>
					<TableHead>Repository</TableHead>
					<TableHead>Forge</TableHead>
					<TableHead>Token</TableHead>
					<TableHead>Status</TableHead>
					<TableHead class="sr-only">Untrack</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each repos as repo (repo.id)}
					<TableRow>
						<TableCell class="font-medium">{repo.identifier}</TableCell>
						<TableCell>{FORGE_LABELS[repo.forge] ?? repo.forge}</TableCell>
						<TableCell>{repo.tokenMasked}</TableCell>
						<TableCell>
							<Badge
								variant={STATUS_VARIANTS[repo.ingestionStatus] ?? 'outline'}
								class={repo.ingestionStatus === 'degraded' ? 'bg-destructive text-white' : ''}
							>
								{repo.ingestionStatus}
							</Badge>
							{#if repo.ingestionStatusReason}
								<span class="ml-1 text-xs text-muted-foreground"
									>{repo.ingestionStatusReason}</span
								>
							{/if}
						</TableCell>
						<TableCell>
							<Button variant="ghost" size="sm" onclick={() => (untrackTarget = repo)}>
								Untrack
							</Button>
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	{/if}
</main>

<AlertDialog open={untrackTarget !== null} onOpenChange={(open) => !open && (untrackTarget = null)}>
	<AlertDialogContent>
		<AlertDialogHeader>
			<AlertDialogTitle>Untrack {untrackTarget?.identifier}?</AlertDialogTitle>
			<AlertDialogDescription>
				This removes the repository and all of its stored runs, jobs, and steps. The
				webhook created on the forge isn't deleted automatically.
			</AlertDialogDescription>
		</AlertDialogHeader>
		{#if untrackError}
			<p role="alert" class="text-sm text-destructive">{untrackError}</p>
		{/if}
		<AlertDialogFooter>
			<AlertDialogCancel disabled={untrackBusy}>Cancel</AlertDialogCancel>
			<AlertDialogAction disabled={untrackBusy} onclick={handleUntrack}>
				{untrackBusy ? 'Untracking…' : 'Untrack'}
			</AlertDialogAction>
		</AlertDialogFooter>
	</AlertDialogContent>
</AlertDialog>
