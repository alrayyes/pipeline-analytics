<script lang="ts">
import { goto } from '$app/navigation';

type Mode = 'checking' | 'register' | 'login' | 'unreachable';

let mode = $state<Mode>('checking');
let loginOptions = $state<PublicKeyCredentialRequestOptionsJSON | null>(null);
let error = $state<string | null>(null);
let busy = $state(false);

async function detectMode(): Promise<void> {
	let res: Response;

	try {
		res = await fetch('/api/auth/login/options', { method: 'POST' });
	} catch {
		mode = 'unreachable';
		return;
	}

	if (res.status === 404) {
		mode = 'register';
		return;
	}

	if (!res.ok) {
		mode = 'unreachable';
		return;
	}

	// go-webauthn wraps its options in { publicKey: ... } to mirror the
	// shape navigator.credentials.get()/create() themselves take, so the
	// client can pass the wrapper straight through -- but
	// parseRequestOptionsFromJSON/parseCreationOptionsFromJSON want only
	// the inner object.
	const body: { publicKey: PublicKeyCredentialRequestOptionsJSON } =
		await res.json();
	loginOptions = body.publicKey;
	mode = 'login';
}

detectMode();

function describeError(err: unknown, fallback: string): string {
	if (err instanceof DOMException && err.name === 'NotAllowedError') {
		return 'Passkey ceremony was cancelled or timed out.';
	}

	return err instanceof Error ? err.message : fallback;
}

async function handleRegister(): Promise<void> {
	busy = true;
	error = null;

	try {
		const optionsRes = await fetch('/api/auth/register/options', {
			method: 'POST',
		});
		if (!optionsRes.ok) {
			throw new Error(
				optionsRes.status === 409
					? 'An account already exists.'
					: 'Could not start registration.',
			);
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

		const finishRes = await fetch('/api/auth/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(credential.toJSON()),
		});
		if (!finishRes.ok) {
			throw new Error('Registration failed.');
		}

		await goto('/');
	} catch (err) {
		error = describeError(err, 'Registration failed.');
	} finally {
		busy = false;
	}
}

async function handleLogin(): Promise<void> {
	if (!loginOptions) {
		return;
	}

	busy = true;
	error = null;

	try {
		const options =
			PublicKeyCredential.parseRequestOptionsFromJSON(loginOptions);
		const credential = await navigator.credentials.get({ publicKey: options });

		if (!(credential instanceof PublicKeyCredential)) {
			throw new Error('Login was not completed.');
		}

		const res = await fetch('/api/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(credential.toJSON()),
		});
		if (!res.ok) {
			throw new Error('Login failed.');
		}

		await goto('/');
	} catch (err) {
		error = describeError(err, 'Login failed.');
	} finally {
		busy = false;
	}
}
</script>

<svelte:head>
	<title>Log in · pipeline-analytics</title>
</svelte:head>

<main>
	<h1>pipeline-analytics</h1>

	{#if mode === 'checking'}
		<p>Checking account status…</p>
	{:else if mode === 'unreachable'}
		<p role="alert">Could not reach the server. Try reloading the page.</p>
	{:else if mode === 'register'}
		<p>No account has been registered yet. Register a passkey to get started.</p>
		<button type="button" onclick={handleRegister} disabled={busy}>
			{busy ? 'Registering…' : 'Register your passkey'}
		</button>
	{:else if mode === 'login'}
		<button type="button" onclick={handleLogin} disabled={busy}>
			{busy ? 'Logging in…' : 'Log in with passkey'}
		</button>
	{/if}

	<p role="alert" aria-live="polite">{error ?? ''}</p>
</main>
