<script lang="ts">
import GitBranchIcon from '@lucide/svelte/icons/git-branch';
import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
import PlayIcon from '@lucide/svelte/icons/play';
import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
import { page } from '$app/state';
import { isActivePath } from '$lib/components/navActive.js';
import { cn } from '$lib/utils.js';

let { hasRepos }: { hasRepos: boolean } = $props();

const tabs = [
	{ href: '/', label: 'Overview', icon: LayoutDashboardIcon },
	{ href: '/pipelines', label: 'Pipelines', icon: GitBranchIcon },
	{ href: '/runs', label: 'Runs', icon: PlayIcon },
	{ href: '/failures', label: 'Failures', icon: TriangleAlertIcon },
];

const base =
	'flex min-h-12 flex-col items-center justify-center gap-0.5 border-t-2 text-xs';
</script>

<nav
	aria-label="Main views"
	class="fixed inset-x-0 bottom-0 border-t bg-background pb-[env(safe-area-inset-bottom)] sm:hidden"
>
	<ul class="grid grid-cols-4">
		{#each tabs as tab (tab.href)}
			{@const active = isActivePath(page.url.pathname, tab.href)}
			<li class="min-w-0">
				{#if hasRepos || tab.href === '/'}
					<a
						href={tab.href}
						class={cn(
							base,
							'border-transparent text-muted-foreground hover:text-foreground',
							active && 'border-foreground font-medium text-foreground',
						)}
						aria-current={active ? 'page' : undefined}
					>
						<tab.icon class="size-5" aria-hidden="true" />
						{tab.label}
					</a>
				{:else}
					<span
						aria-disabled="true"
						role="link"
						tabindex="-1"
						class={cn(base, 'cursor-not-allowed border-transparent text-muted-foreground/50')}
					>
						<tab.icon class="size-5" aria-hidden="true" />
						{tab.label}
					</span>
				{/if}
			</li>
		{/each}
	</ul>
</nav>
