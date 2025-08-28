<script lang="ts">
    import { cn } from "$lib/utils";
    import type { HTMLAttributes } from "svelte/elements";

    interface $$Props extends HTMLAttributes<HTMLDivElement> {
        class?: string;
        orientation?: "vertical" | "horizontal" | "both";
    }

    let className: string | undefined = undefined;
    export { className as class };
    export let orientation: "vertical" | "horizontal" | "both" = "vertical";

    $: classes = cn("relative overflow-hidden", className);

    $: scrollbarClasses = cn(
        "scrollbar-thin scrollbar-track-transparent",
        orientation === "vertical" &&
            "scrollbar-thumb-border hover:scrollbar-thumb-border/80",
        orientation === "horizontal" &&
            "scrollbar-thumb-border hover:scrollbar-thumb-border/80",
        orientation === "both" &&
            "scrollbar-thumb-border hover:scrollbar-thumb-border/80",
    );

    $: contentClasses = cn(
        "h-full w-full rounded-[inherit]",
        orientation === "vertical" && "overflow-y-auto overflow-x-hidden",
        orientation === "horizontal" && "overflow-x-auto overflow-y-hidden",
        orientation === "both" && "overflow-auto",
        scrollbarClasses,
    );
</script>

<div class={classes} {...$$restProps}>
    <div class={contentClasses}>
        <slot />
    </div>
</div>

<style>
    /* Custom scrollbar styles */
    :global(.scrollbar-thin) {
        scrollbar-width: thin;
    }

    :global(.scrollbar-thin::-webkit-scrollbar) {
        width: 8px;
        height: 8px;
    }

    :global(.scrollbar-thin::-webkit-scrollbar-track) {
        background: transparent;
    }

    :global(.scrollbar-thin::-webkit-scrollbar-thumb) {
        background-color: hsl(var(--border));
        border-radius: 4px;
        border: 2px solid transparent;
        background-clip: content-box;
    }

    :global(.scrollbar-thin::-webkit-scrollbar-thumb:hover) {
        background-color: hsl(var(--border) / 0.8);
    }

    :global(.scrollbar-thin::-webkit-scrollbar-corner) {
        background: transparent;
    }

    /* Firefox */
    :global(.scrollbar-thin) {
        scrollbar-color: hsl(var(--border)) transparent;
    }

    /* Show scrollbar on hover */
    :global(.scrollbar-thin) {
        scrollbar-width: none;
    }

    :global(.scrollbar-thin:hover) {
        scrollbar-width: thin;
    }

    :global(.scrollbar-thin::-webkit-scrollbar) {
        width: 0px;
        height: 0px;
    }

    :global(.scrollbar-thin:hover::-webkit-scrollbar) {
        width: 8px;
        height: 8px;
    }

    /* Always show scrollbar variant */
    :global(.scrollbar-always) {
        scrollbar-width: thin;
    }

    :global(.scrollbar-always::-webkit-scrollbar) {
        width: 8px;
        height: 8px;
    }

    :global(.scrollbar-always::-webkit-scrollbar-track) {
        background: hsl(var(--muted) / 0.1);
    }

    :global(.scrollbar-always::-webkit-scrollbar-thumb) {
        background-color: hsl(var(--border));
        border-radius: 4px;
    }

    :global(.scrollbar-always::-webkit-scrollbar-thumb:hover) {
        background-color: hsl(var(--border) / 0.8);
    }
</style>
