<script lang="ts">
    import { cn } from "$lib/utils";
    import type { Snippet } from "svelte";
    import { fly } from "svelte/transition";

    interface Props {
        content: string;
        side?: "top" | "bottom" | "left" | "right";
        class?: string;
        children: Snippet;
    }

    let { content, side = "top", class: className, children }: Props = $props();

    let isVisible = $state(false);
    let tooltipElement: HTMLDivElement | undefined = $state();
    let triggerElement: HTMLDivElement | undefined = $state();

    function showTooltip() {
        isVisible = true;
    }

    function hideTooltip() {
        isVisible = false;
    }

    const sideClasses = {
        top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
        bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
        left: "right-full top-1/2 -translate-y-1/2 mr-2",
        right: "left-full top-1/2 -translate-y-1/2 ml-2",
    };

    const arrowClasses = {
        top: "top-full left-1/2 -translate-x-1/2 border-l-transparent border-r-transparent border-b-transparent",
        bottom: "bottom-full left-1/2 -translate-x-1/2 border-l-transparent border-r-transparent border-t-transparent",
        left: "left-full top-1/2 -translate-y-1/2 border-t-transparent border-b-transparent border-r-transparent",
        right: "right-full top-1/2 -translate-y-1/2 border-t-transparent border-b-transparent border-l-transparent",
    };

    const transitionConfig = {
        top: { y: 5, duration: 150 },
        bottom: { y: -5, duration: 150 },
        left: { x: 5, duration: 150 },
        right: { x: -5, duration: 150 },
    };
</script>

<div class="relative inline-block" bind:this={triggerElement}>
    <div
        role="button"
        tabindex="0"
        onmouseenter={showTooltip}
        onmouseleave={hideTooltip}
        onfocus={showTooltip}
        onblur={hideTooltip}
        class={cn("inline-flex", className)}
    >
        {@render children()}
    </div>

    {#if isVisible}
        <div
            bind:this={tooltipElement}
            role="tooltip"
            transition:fly={transitionConfig[side]}
            class={cn(
                "absolute z-50 px-2 py-1 text-[11px] font-medium bg-foreground text-background border border-foreground rounded shadow-lg whitespace-nowrap pointer-events-none",
                sideClasses[side],
            )}
        >
            {content}
            <div
                class={cn(
                    "absolute w-0 h-0 border-4 border-foreground",
                    arrowClasses[side],
                )}
            ></div>
        </div>
    {/if}
</div>
