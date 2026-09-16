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
let registerError = $state<string | null>(null);
let forge = $state<'github' | 'forgejo'>('github');
let identifier = $state('');
let forgejoInstanceUrl = $state('');
let token = $state('');

let discoveredRepos = $state<string[] | null>(null);
let discoverBusy = $state(false);
let discoverError = $state<string | null>(null);
let manualEntry = $state(false);

const canDiscover = $derived(
	token.trim() !== '' &&
		(forge === 'github' || forgejoInstanceUrl.trim() !== ''),
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
// fetched for -- invalidate it the moment any of those change underneath
// it, rather than letting a stale picker suggest repos the current token
// might not even reach.
$effect(() => {
	[forge, token, forgejoInstanceUrl];
	discoveredRepos = null;
	discoverError = null;
	manualEntry = false;
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
	forge = 'github';
	identifier = '';
	forgejoInstanceUrl = '';
	token = '';
	registerError = null;
	discoveredRepos = null;
	discoverError = null;
	manualEntry = false;
}

async function handleDiscover(): Promise<void> {
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

		discoveredRepos = await res.json();
		identifier = discoveredRepos?.[0] ?? '';
	} catch {
		discoverError = 'Could not reach the server.';
	} finally {
		discoverBusy = false;
	}
}

async function handleRegister(event: SubmitEvent): Promise<void> {
	event.preventDefault();
	registerBusy = true;
	registerError = null;

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

		if (!res.ok) {
			const body = await res.json().catch(() => null);
			registerError = body?.message ?? 'Could not register the repository.';

			return;
		}

		rememberToken(
			forge,
			forge === 'forgejo' ? forgejoInstanceUrl : undefined,
			token,
		);

		registerOpen = false;
		resetForm();
		await loadRepos();
	} catch {
		registerError = 'Could not reach the server.';
	} finally {
		registerBusy = false;
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
		await loadRepos();
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
				<form onsubmit={handleRegister}>
					<DialogHeader>
						<DialogTitle>Register a repository</DialogTitle>
						<DialogDescription>
							Starts tracking a GitHub or Forgejo repository's Actions runs.
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
									onclick={() => (token = rememberedForCurrentForge ?? '')}
								>
									Use saved token ({maskToken(rememberedForCurrentForge)})
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

						<div class="grid gap-2">
							<div class="flex items-center justify-between">
								<Label for="identifier">Repository</Label>
								{#if discoveredRepos !== null && !manualEntry}
									<button
										type="button"
										class="text-xs text-muted-foreground underline hover:text-foreground"
										onclick={() => (manualEntry = true)}
									>
										Enter manually instead
									</button>
								{/if}
							</div>

							{#if discoveredRepos !== null && !manualEntry}
								{#if discoveredRepos.length === 0}
									<p class="text-sm text-muted-foreground">
										That token can't see any repositories.
									</p>
								{:else}
									<Select type="single" bind:value={identifier}>
										<SelectTrigger id="identifier" class="w-full">
											{identifier || 'Select a repository'}
										</SelectTrigger>
										<SelectContent>
											{#each discoveredRepos as repoId (repoId)}
												<SelectItem value={repoId} label={repoId}>{repoId}</SelectItem>
											{/each}
										</SelectContent>
									</Select>
								{/if}
							{:else}
								<div class="flex gap-2">
									<Input
										id="identifier"
										bind:value={identifier}
										placeholder="owner/name"
										required
										class="flex-1"
									/>
									<Button
										type="button"
										variant="outline"
										disabled={!canDiscover || discoverBusy}
										onclick={handleDiscover}
									>
										{discoverBusy ? 'Finding…' : 'Find repositories'}
									</Button>
								</div>
							{/if}

							{#if discoverError}
								<p role="alert" class="text-sm text-destructive">{discoverError}</p>
							{/if}
						</div>

						{#if registerError}
							<p role="alert" class="text-sm text-destructive">{registerError}</p>
						{/if}
					</div>

					<DialogFooter>
						<Button type="submit" disabled={registerBusy}>
							{registerBusy ? 'Registering…' : 'Register'}
						</Button>
					</DialogFooter>
				</form>
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
