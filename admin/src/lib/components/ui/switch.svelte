<script lang="ts">
    import { createEventDispatcher } from "svelte";

    export let checked = false;
    export let disabled = false;
    export let id: string | undefined = undefined;
    export let name: string | undefined = undefined;
    export let value: string | undefined = undefined;

    const dispatch = createEventDispatcher<{
        change: boolean;
    }>();

    function handleClick() {
        if (disabled) return;
        checked = !checked;
        dispatch("change", checked);
    }

    function handleKeydown(event: KeyboardEvent) {
        if (disabled) return;
        if (event.key === " " || event.key === "Enter") {
            event.preventDefault();
            handleClick();
        }
    }
</script>

<button
    {id}
    type="button"
    role="switch"
    aria-checked={checked}
    aria-disabled={disabled}
    {disabled}
    class="peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 {checked
        ? 'bg-primary'
        : 'bg-input'}"
    on:click={handleClick}
    on:keydown={handleKeydown}
>
    <span
        class="pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform {checked
            ? 'translate-x-5'
            : 'translate-x-0'}"
    ></span>
</button>

{#if name}
    <input type="checkbox" {name} {value} bind:checked style="display: none;" />
{/if}
