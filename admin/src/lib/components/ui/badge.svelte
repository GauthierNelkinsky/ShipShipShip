<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

	interface $$Props extends HTMLAttributes<HTMLDivElement> {
		variant?: BadgeVariant;
		class?: string;
	}

	export let variant: BadgeVariant = 'default';
	let className: string | undefined = undefined;
	export { className as class };

	const variants = {
		default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
		secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
		destructive: 'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
		outline: 'text-foreground'
	};

	$: classes = cn(
		'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
		variants[variant],
		className
	);
</script>

<div class={classes} {...$$restProps}>
	<slot />
</div>
