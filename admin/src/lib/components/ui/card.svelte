<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';

	interface $$Props extends HTMLAttributes<HTMLDivElement> {
		class?: string;
	}

	let className: string | undefined = undefined;
	export { className as class };

	$: classes = cn(
		'rounded-lg border bg-card text-card-foreground shadow-sm',
		className
	);
</script>

<div class={classes} {...$$restProps}>
	<slot />
</div>
