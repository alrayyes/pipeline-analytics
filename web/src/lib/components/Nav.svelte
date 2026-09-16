<script lang="ts">
import { goto } from '$app/navigation';
import { page } from '$app/state';
import { Button } from '$lib/components/ui/button/index.js';
import { cn } from '$lib/utils.js';

let logoutBusy = $state(false);

const links = [
	{ href: '/', label: 'Pipelines' },
	{ href: '/repos', label: 'Repositories' },
];

function isActive(href: string): boolean {
	return href === '/'
		? page.url.pathname === '/'
		: page.url.pathname.startsWith(href);
}

async function handleLogout(): Promise<void> {
	logoutBusy = true;

	try {
		await fetch('/api/auth/logout', { method: 'POST' });
	} finally {
		await goto('/login');
	}
}
</script>

<header class="border-b">
	<div class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3">
		<nav class="flex items-center gap-6">
			<span class="font-semibold">pipeline-analytics</span>
			<ul class="flex items-center gap-4 text-sm">
				{#each links as link (link.href)}
					<li>
						<a
							href={link.href}
							class={cn(
								'text-muted-foreground hover:text-foreground',
								isActive(link.href) && 'font-medium text-foreground',
							)}
							aria-current={isActive(link.href) ? 'page' : undefined}
						>
							{link.label}
						</a>
					</li>
				{/each}
			</ul>
		</nav>
		<Button variant="outline" size="sm" onclick={handleLogout} disabled={logoutBusy}>
			{logoutBusy ? 'Logging out…' : 'Log out'}
		</Button>
	</div>
</header>
