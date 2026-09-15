<script lang="ts">
import { goto } from '$app/navigation';

let busy = $state(false);

async function handleLogout(): Promise<void> {
	busy = true;

	try {
		await fetch('/api/auth/logout', { method: 'POST' });
	} finally {
		await goto('/login');
	}
}
</script>

<svelte:head>
	<title>pipeline-analytics</title>
</svelte:head>

<main>
	<h1>pipeline-analytics</h1>
	<p>
		Self-hosted pipeline analytics for GitHub Actions and Forgejo Actions. The pipeline overview
		isn't built yet.
	</p>
	<button type="button" onclick={handleLogout} disabled={busy}>
		{busy ? 'Logging out…' : 'Log out'}
	</button>
</main>
