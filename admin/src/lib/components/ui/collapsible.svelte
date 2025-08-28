<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { slide } from "svelte/transition";
    import { quintOut } from "svelte/easing";

    export let open = false;
    export let disabled = false;

    const dispatch = createEventDispatcher<{
        openChange: boolean;
    }>();

    function toggle() {
        if (disabled) return;
        open = !open;
        dispatch("openChange", open);
    }

    export { toggle };
</script>

<div
    class="collapsible"
    data-state={open ? "open" : "closed"}
    data-disabled={disabled}
>
    <div class="collapsible-trigger">
        <slot name="trigger" {toggle} {open} />
    </div>
    {#if open}
        <div
            class="collapsible-content"
            transition:slide={{ duration: 200, easing: quintOut }}
        >
            <slot />
        </div>
    {/if}
</div>

<style>
    .collapsible {
        width: 100%;
    }

    .collapsible-content {
        overflow: hidden;
    }
</style>
