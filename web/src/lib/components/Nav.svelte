<script lang="ts">
import MenuIcon from '@lucide/svelte/icons/menu';
import SettingsIcon from '@lucide/svelte/icons/settings';
import XIcon from '@lucide/svelte/icons/x';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import logo from '$lib/assets/favicon.svg';
import { Button } from '$lib/components/ui/button/index.js';
import { cn } from '$lib/utils.js';

let { hasRepos }: { hasRepos: boolean } = $props();

let logoutBusy = $state(false);

// The four nav links plus the "Register a repository" CTA don't fit a
// phone-width header in one row (#171) -- collapsed behind this toggle
// there instead of wrapping, which the outer header row already does for
// its own two direct children (this nav vs. the theme/logout controls) but
// can't do for what's inside a single flex child.
let mobileMenuOpen = $state(false);

// A real navigation should always leave the menu closed behind it, rather
// than the panel still open over whatever page it lands the reader on.
$effect(() => {
	page.url.pathname;
	mobileMenuOpen = false;
});

const links = [
	{ href: '/', label: 'Pipelines' },
	{ href: '/steps', label: 'Steps' },
	{ href: '/repos', label: 'Repositories' },
	{ href: '/insights', label: 'Insights' },
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

{#snippet navLinks()}
	<ul class="flex items-center gap-4 text-sm max-sm:flex-col max-sm:items-start">
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
{/snippet}

<header class="border-b">
	<div class="mx-auto flex max-w-4xl flex-wrap items-center justify-between gap-y-2 px-4 py-3">
		<div class="flex items-center gap-6">
			<a href="/" aria-label="Pipeline Analytics home" class="flex items-center">
				<img src={logo} alt="" class="size-6" />
			</a>
			<nav class="hidden items-center gap-6 sm:flex">
				{@render navLinks()}
			</nav>
		</div>
		<div class="flex items-center gap-2">
			<Button
				variant="outline"
				size="icon"
				href="/settings"
				aria-label="Settings"
				title="Settings"
				aria-current={page.url.pathname === '/settings' ? 'page' : undefined}
			>
				<SettingsIcon />
			</Button>
			<Button variant="outline" size="sm" onclick={handleLogout} disabled={logoutBusy}>
				{logoutBusy ? 'Logging out…' : 'Log out'}
			</Button>
			<Button
				variant="outline"
				size="icon"
				class="sm:hidden"
				aria-expanded={mobileMenuOpen}
				aria-controls="mobile-nav"
				aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
				onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
			>
				{#if mobileMenuOpen}
					<XIcon />
				{:else}
					<MenuIcon />
				{/if}
			</Button>
		</div>
	</div>
	{#if mobileMenuOpen}
		<nav id="mobile-nav" class="flex flex-col items-start gap-3 border-t px-4 py-3 sm:hidden">
			{@render navLinks()}
		</nav>
	{/if}
</header>
