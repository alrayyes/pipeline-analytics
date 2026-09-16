<script lang="ts">
import MonitorIcon from '@lucide/svelte/icons/monitor';
import MoonIcon from '@lucide/svelte/icons/moon';
import SunIcon from '@lucide/svelte/icons/sun';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import logo from '$lib/assets/favicon.svg';
import { Button } from '$lib/components/ui/button/index.js';
import { cycleTheme, getTheme } from '$lib/theme.svelte.js';
import { cn } from '$lib/utils.js';

let { hasRepos }: { hasRepos: boolean } = $props();

let logoutBusy = $state(false);

const THEME_LABELS = {
	light: 'Light',
	dark: 'Dark',
	system: 'System',
} as const;

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
	<div class="mx-auto flex max-w-4xl flex-wrap items-center justify-between gap-y-2 px-4 py-3">
		<nav class="flex items-center gap-6">
			<a href="/" aria-label="Pipeline Analytics home" class="flex items-center">
				<img src={logo} alt="" class="size-6" />
			</a>
			<ul class="flex items-center gap-4 text-sm">
				{#each links as link (link.href)}
					<li>
						{#if hasRepos}
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
						{:else}
							<span
								aria-disabled="true"
								role="link"
								tabindex="-1"
								class="cursor-not-allowed text-muted-foreground/50"
							>
								{link.label}
							</span>
						{/if}
					</li>
				{/each}
			</ul>
			{#if !hasRepos}
				<Button href="/repos" size="sm">Register a repository</Button>
			{/if}
		</nav>
		<div class="flex items-center gap-2">
			<Button
				variant="outline"
				size="icon"
				onclick={cycleTheme}
				aria-label="Theme: {THEME_LABELS[getTheme()]}. Click to change."
				title="Theme: {THEME_LABELS[getTheme()]}. Click to change."
			>
				{#if getTheme() === 'light'}
					<SunIcon />
				{:else if getTheme() === 'dark'}
					<MoonIcon />
				{:else}
					<MonitorIcon />
				{/if}
			</Button>
			<Button variant="outline" size="sm" onclick={handleLogout} disabled={logoutBusy}>
				{logoutBusy ? 'Logging out…' : 'Log out'}
			</Button>
		</div>
	</div>
</header>
