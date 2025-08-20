<script lang="ts">
    import { Toaster as SonnerPrimitive } from "svelte-sonner";
    import { theme } from "$lib/stores/theme";

    export let closeButton = false;
    export let richColors = false;
    export let expand = false;
    export let visibleToasts = 1;
    export let position:
        | "top-left"
        | "top-right"
        | "bottom-left"
        | "bottom-right"
        | "top-center"
        | "bottom-center" = "top-right";
    export let offset = "16px";
    export let dir: "auto" | "ltr" | "rtl" = "auto";
    export let hotkey = ["altKey", "KeyT"];
    export let toastOptions = {
        style: "min-width: 320px; max-width: 420px;",
    };
    export let duration = 4000;
</script>

<SonnerPrimitive
    theme={$theme}
    {closeButton}
    {richColors}
    {expand}
    {visibleToasts}
    {position}
    {offset}
    {dir}
    {hotkey}
    {duration}
    {toastOptions}
    {...$$restProps}
/>
