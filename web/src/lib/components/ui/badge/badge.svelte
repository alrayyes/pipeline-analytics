<script lang="ts" module>
import { tv, type VariantProps } from 'tailwind-variants';

export const badgeVariants = tv({
	base: 'h-5 gap-1 rounded-4xl border border-transparent px-2 py-0.5 text-xs font-medium transition-all has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&>svg]:size-3! group/badge inline-flex w-fit shrink-0 items-center justify-center overflow-hidden whitespace-nowrap focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&>svg]:pointer-events-none',
	variants: {
		variant: {
			default: 'bg-primary text-primary-foreground [a]:hover:bg-primary/80',
			secondary:
				'bg-secondary text-secondary-foreground [a]:hover:bg-secondary/80',
			destructive:
				'bg-destructive/10 [a]:hover:bg-destructive/20 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 text-destructive dark:bg-destructive/20',
			// text-green-700 measured 4.41:1 against the composited badge
			// background -- short of WCAG AA's 4.5:1 floor -- caught live by
			// axe-core on the first page that actually rendered it as text.
			// Darkened to green-800, which clears it with room to spare.
			success:
				'bg-green-600/10 text-green-800 [a]:hover:bg-green-600/20 dark:bg-green-500/20 dark:text-green-400',
			outline:
				'border-border text-foreground [a]:hover:bg-muted [a]:hover:text-muted-foreground',
			ghost:
				'hover:bg-muted hover:text-muted-foreground dark:hover:bg-muted/50',
			link: 'text-primary underline-offset-4 hover:underline',
		},
	},
	defaultVariants: {
		variant: 'default',
	},
});

export type BadgeVariant = VariantProps<typeof badgeVariants>['variant'];
</script>

<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAnchorAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		href,
		class: className,
		variant = "default",
		children,
		...restProps
	}: WithElementRef<HTMLAnchorAttributes> & {
		variant?: BadgeVariant;
	} = $props();
</script>

<svelte:element
	this={href ? "a" : "span"}
	bind:this={ref}
	data-slot="badge"
	{href}
	class={cn(badgeVariants({ variant }), className)}
	{...restProps}
>
	{@render children?.()}
</svelte:element>
