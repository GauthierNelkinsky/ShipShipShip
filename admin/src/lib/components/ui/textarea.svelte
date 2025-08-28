<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLTextareaAttributes } from 'svelte/elements';

	interface $$Props extends HTMLTextareaAttributes {
		class?: string;
		value?: string;
	}

	let className: string | undefined = undefined;
	export { className as class };
	export let value: string = '';

	$: classes = cn(
		'flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
		className
	);
</script>

<textarea
	bind:value
	class={classes}
	{...$$restProps}
	on:blur
	on:change
	on:click
	on:focus
	on:keydown
	on:keypress
	on:keyup
	on:mouseover
	on:mouseenter
	on:mouseleave
	on:paste
	on:input
></textarea>
